package database

import (
	"context"
	"fmt"
	"github.com/hsmade/OSM-ARDF/pkg/measurement"
	"github.com/jackc/pgx/v4"
	"github.com/ory/dockertest"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

var (
	dbResource *dockertest.Resource
	dockerPort int
)

// Setup a docker container with postgres before running the tests for databases
func TestMain(m *testing.M) {
	var db *pgx.Conn
	var err error
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	initScript, err := filepath.Abs(dir + "../../../scripts/init-gis.sh")
	if err != nil {
		log.Fatal(err)
	}

	options := &dockertest.RunOptions{
		Repository: "timescale/timescaledb-postgis",
		Tag:        "1.4.2-pg11",
		Env:        []string{"POSTGRES_PASSWORD=postgres"},
		Mounts:     []string{initScript + ":/docker-entrypoint-initdb.d/init-gis.sh"},
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
	testMeasurement := measurement.Measurement{
		Timestamp: time.Now(),
		Station:   "test",
		Longitude: 1,
		Latitude:  2,
		Bearing:   3,
	}

	type fields struct {
		Host         string
		Port         uint16
		Username     string
		Password     string
		DatabaseName string
	}
	type args struct {
		m *measurement.Measurement
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Happy path",
			fields{
				Host:         "localhost",
				Port:         uint16(dockerPort),
				Username:     "postgres",
				Password:     "postgres",
				DatabaseName: "postgres",
			},
			args{
				&testMeasurement,
			},
			false,
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
			if err := d.Add(tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
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
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Happy path",
			fields: fields{
				Host:         "localhost",
				Port:         uint16(dockerPort),
				Username:     "postgres",
				Password:     "postgres",
				DatabaseName: "postgres",
			},
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
