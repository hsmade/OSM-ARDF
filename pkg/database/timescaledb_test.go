package database

import (
	"context"
	"fmt"
	"github.com/hsmade/OSM-ARDF/pkg/types"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ory/dockertest"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkb"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"
)

var (
	dbResource *dockertest.Resource
	dockerPort int
	db         *pgx.Conn
)

// Setup a docker container with postgres before running the tests for databases
func TestMain(m *testing.M) {
	var err error
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	initSql, err := filepath.Abs(dir + "../../../scripts/init-gis.sql")
	if err != nil {
		log.Fatal(err)
	}

	options := &dockertest.RunOptions{
		Repository: "timescale/timescaledb-postgis",
		Tag:        "1.4.2-pg11",
		Env:        []string{"POSTGRES_PASSWORD=postgres"},
		Mounts:     []string{initSql + ":/docker-entrypoint-initdb.d/init-gis.sql"},
	}
	dbResource, err = pool.RunWithOptions(options)
	if err != nil {
		log.Fatalf("Could not start dbResource: %s", err)
	}
	err = dbResource.Expire(60)
	if err != nil {
		log.Fatalf("Could not set expiration for docker container: %s", err)
	}

	if err = pool.Retry(func() error {
		var err error
		db, err = pgx.Connect(
			context.Background(),
			fmt.Sprintf("postgresql://postgres:postgres@localhost:%s/postgres?sslmode=disable", dbResource.GetPort("5432/tcp")))
		if err != nil {
			return err
		}

		_, err = db.Exec(context.Background(), "SELECT * FROM doppler;")
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	dockerPort, err = strconv.Atoi(dbResource.GetPort("5432/tcp"))
	if err != nil {
		log.Fatalf("Failed to get port for docker container: %e", err)
	}

	exit := m.Run()

	err = pool.Purge(dbResource)
	os.Exit(exit)
}

func TestTimescaleDB_Add(t *testing.T) {
	timeNow := time.Now().Truncate(time.Second)

	type fields struct {
		Host         string
		Port         uint16
		Username     string
		Password     string
		DatabaseName string
	}

	type args struct {
		m *types.Measurement
	}

	testDBFields := fields{
		Host:         "localhost",
		Port:         uint16(dockerPort),
		Username:     "postgres",
		Password:     "postgres",
		DatabaseName: "postgres",
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    args
		wantErr bool
	}{
		{
			"Happy path",
			testDBFields,
			args{&types.Measurement{
				Timestamp: timeNow,
				Station:   "test_Add_happy_path",
				Longitude: 1,
				Latitude:  2,
				Bearing:   3,
			}},
			args{&types.Measurement{
				Timestamp: timeNow,
				Station:   "test_Add_happy_path",
				Longitude: 1,
				Latitude:  2,
				Bearing:   3,
			}},
			false,
		},
		{
			"negative bearing",
			testDBFields,
			args{&types.Measurement{
				Timestamp: timeNow,
				Bearing:   -1,
			}},
			args{nil},
			true,
		},
		{
			"too high bearing",
			testDBFields,
			args{&types.Measurement{
				Timestamp: timeNow,
				Bearing:   361,
			}},
			args{nil},
			true,
		},
		{
			"no station name",
			testDBFields,
			args{&types.Measurement{
				Timestamp: timeNow,
				Station:   "",
			}},
			args{nil},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &TimescaleDB{
				Host:         tt.fields.Host,
				Port:         tt.fields.Port,
				Username:     tt.fields.Username,
				Password:     tt.fields.Password,
				DatabaseName: tt.fields.DatabaseName,
			}
			err := d.Connect()
			if err != nil {
				t.Errorf("Failed during Connect(): %e", err)
			}

			err = d.Add(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				query := fmt.Sprintf("SELECT time, station, ST_AsBinary(point), ST_AsBinary(line), bearing FROM doppler WHERE station = '%s'", tt.want.m.Station)
				row := db.QueryRow(context.Background(), query)

				var got types.Measurement
				var point orb.Point
				var line orb.LineString
				err = row.Scan(&got.Timestamp, &got.Station, wkb.Scanner(&point), wkb.Scanner(&line), &got.Bearing)

				if point.X() != line[0].X() || point.Y() != line[0].Y() {
					t.Errorf("start of line: %v doesn't match point: %v", line[0], point)
				}
				got.Longitude = point.X()
				got.Latitude = point.Y()
				if err != nil {
					t.Fatalf("failed to parse row: %e", err)
				}

				if !reflect.DeepEqual(*tt.want.m, got) {
					t.Errorf("Measurement changed during storage: %v, got: %v", *tt.want.m, got)
				}
			}
		})
	}
}

func TestTimescaleDB_Connect(t *testing.T) {
	type fields struct {
		Host         string
		Port         uint16
		Username     string
		Password     string
		DatabaseName string
	}

	testDBFields := fields{
		Host:         "localhost",
		Port:         uint16(dockerPort),
		Username:     "postgres",
		Password:     "postgres",
		DatabaseName: "postgres",
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "Happy path",
			fields:  testDBFields,
			wantErr: false,
		},
		{
			name: "Connection failed",
			fields: fields{
				Host:     "900.900.900.900",
				Port:     0,
				Username: "postgres",
				Password: "postgres",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &TimescaleDB{
				Host:     tt.fields.Host,
				Port:     tt.fields.Port,
				Username: tt.fields.Username,
				Password: tt.fields.Password,
			}
			if err := d.Connect(); (err != nil) != tt.wantErr {
				t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewTimescaleDB_HappyPath(t *testing.T) {
	databaseURL := "postgresql://user:pass@host:1234/db?sslmode=disable"
	db := New(databaseURL)
	want := TimescaleDB{
		Host:           "host",
		Port:           1234,
		Username:       "user",
		Password:       "pass",
		DatabaseName:   "db",
		connectionPool: nil,
	}

	if !reflect.DeepEqual(&want, db) {
		t.Errorf("invalid db object created. Wanted: %v, got: %v", want, *db)
	}
}

func TestNewTimescaleDB_InvalidURL(t *testing.T) {
	databaseURL := "someinnvalidurl"
	db := New(databaseURL)

	if db != nil {
		t.Errorf("expected nil, got: %v", db)
	}
}

func TestTimescaleDB_GetPositions(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	testMeasurement := types.Measurement{
		Timestamp: now,
		Station:   "test",
		Longitude: 1,
		Latitude:  2,
		Bearing:   3,
	}

	type fields struct {
		Host           string
		Port           uint16
		Username       string
		Password       string
		DatabaseName   string
		connectionPool *pgxpool.Pool
	}

	testDBFields := fields{
		Host:         "localhost",
		Port:         uint16(dockerPort),
		Username:     "postgres",
		Password:     "postgres",
		DatabaseName: "postgres",
	}

	tests := []struct {
		name   string
		fields fields
		input  []*types.Measurement
		want   []*types.Position
	}{
		{
			name:   "single measurement",
			fields: testDBFields,
			input:  []*types.Measurement{&testMeasurement},
			want: []*types.Position{{
				Timestamp: now,
				Station:   "test",
				Longitude: 1,
				Latitude:  2,
			}},
		},
		{
			name:   "two measurements",
			fields: testDBFields,
			input: []*types.Measurement{
				{
					Timestamp: now,
					Station:   "test1",
					Bearing:   1,
				},
				{
					Timestamp: now,
					Station:   "test2",
					Bearing:   1,
				},
			},
			want: []*types.Position{
				{
					Timestamp: now,
					Station:   "test1",
				},
				{
					Timestamp: now,
					Station:   "test2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &TimescaleDB{
				Host:           tt.fields.Host,
				Port:           tt.fields.Port,
				Username:       tt.fields.Username,
				Password:       tt.fields.Password,
				DatabaseName:   tt.fields.DatabaseName,
				connectionPool: tt.fields.connectionPool,
			}
			err := d.Connect()
			if err != nil {
				t.Fatalf("failed to connect to database: %e", err)
			}

			for _, measurement := range tt.input {
				err = d.Add(measurement)
				if err != nil {
					t.Fatalf("failed to insert test measurement: %e", err)
				}
			}

			positions, err := d.GetPositions(1 * time.Minute)
			if err != nil {
				t.Fatalf("failed to query for positions: %e", err)
			}

			for _, want := range tt.want {
				found := false
				for _, position := range positions {
					if reflect.DeepEqual(position, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("could not find item %v in list %v", want, positions[0])
				}
			}
		})
	}
}
func TestTimescaleDB_GetLines(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	testMeasurement := types.Measurement{
		Timestamp: now,
		Station:   "test",
		Longitude: 1,
		Latitude:  2,
		Bearing:   180,
	}

	type fields struct {
		Host           string
		Port           uint16
		Username       string
		Password       string
		DatabaseName   string
		connectionPool *pgxpool.Pool
	}

	testDBFields := fields{
		Host:         "localhost",
		Port:         uint16(dockerPort),
		Username:     "postgres",
		Password:     "postgres",
		DatabaseName: "postgres",
	}

	tests := []struct {
		name   string
		fields fields
		input  []*types.Measurement
		want   []*types.Line
	}{
		{
			name:   "single measurement",
			fields: testDBFields,
			input:  []*types.Measurement{&testMeasurement},
			want: []*types.Line{{
				Position: types.Position{
					Timestamp: now,
					Station:   "test",
					Longitude: 1,
					Latitude:  2,
				},
				LongitudeEnd: 1.0000000000000395,
				LatitudeEnd: 1.910067839408127,
			}},
		},
		{
			name:   "two measurements",
			fields: testDBFields,
			input: []*types.Measurement{
				{
					Timestamp: now,
					Station:   "test1",
					Bearing:   1,
				},
				{
					Timestamp: now,
					Station:   "test2",
					Bearing:   2,
				},
			},
			want: []*types.Line{
				{
					Position: types.Position{
						Timestamp: now,
						Station:   "test1",
					},
					LongitudeEnd: 0.0015695339070129915,
					LatitudeEnd: 0.08991846347697284,
				},
				{
					Position: types.Position{
						Timestamp: now,
						Station:   "test2",
					},
					LongitudeEnd: 0.0031385897164049313,
					LatitudeEnd:0.08987737630457687,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &TimescaleDB{
				Host:           tt.fields.Host,
				Port:           tt.fields.Port,
				Username:       tt.fields.Username,
				Password:       tt.fields.Password,
				DatabaseName:   tt.fields.DatabaseName,
				connectionPool: tt.fields.connectionPool,
			}
			err := d.Connect()
			if err != nil {
				t.Fatalf("failed to connect to database: %e", err)
			}

			for _, measurement := range tt.input {
				err = d.Add(measurement)
				if err != nil {
					t.Fatalf("failed to insert test measurement: %e", err)
				}
			}

			lines, err := d.GetLines(1 * time.Minute)
			if err != nil {
				t.Fatalf("failed to query for lines: %e", err)
			}

			for _, want := range tt.want {
				found := false
				for _, line := range lines {
					if reflect.DeepEqual(line, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("could not find item %v in list %v", want, lines[0])
				}
			}
		})
	}
}
