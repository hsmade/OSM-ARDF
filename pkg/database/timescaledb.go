package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/apex/log"
	"github.com/hsmade/OSM-ARDF/pkg/types"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/kellydunn/golang-geo"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkb"
	"github.com/xo/dburl"
	"strconv"
	"strings"
	"time"
)

type TimescaleDB struct {
	Host           string
	Port           uint16
	Username       string
	Password       string
	DatabaseName   string
	connectionPool *pgxpool.Pool
}

func New(databaseURL string) *TimescaleDB {
	url, err := dburl.Parse(databaseURL)
	if err != nil {
		log.Errorf("failed to parse database url (%s): %e", databaseURL, err)
		return nil
	}

	port, err := strconv.Atoi(url.Port())
	if err != nil {
		log.Errorf("failed to parse port in database url (%s): %e", databaseURL, err)
		return nil
	}

	password, _ := url.User.Password()

	return &TimescaleDB{
		Host:         url.Hostname(),
		Port:         uint16(port),
		Username:     url.User.Username(),
		Password:     password,
		DatabaseName: strings.TrimPrefix(url.Path, "/"),
	}
}

func (d *TimescaleDB) Connect() error {
	url := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		d.Username,
		d.Password,
		d.Host,
		d.Port,
		d.DatabaseName,
	)
	pool, err := pgxpool.Connect(context.Background(), url)
	d.connectionPool = pool
	return err
}

func (d *TimescaleDB) Add(m *types.Measurement) error {
	if m.Bearing > 360 || m.Bearing < 0 {
		return errors.New("bearing must be 0 - 360")
	}

	if m.Station == "" {
		return errors.New("missing station name")
	}

	if d.connectionPool == nil {
		return errors.New("please connect to the database first")
	}
	conn, err := d.connectionPool.Acquire(context.Background())
	if err != nil {
		return err
	}

	defer conn.Release()

	startPoint := geo.NewPoint(m.Latitude, m.Longitude)
	endPoint := startPoint.PointAtDistanceAndBearing(25, float64(m.Bearing))

	query := "insert into \"doppler\"(time, station, point, line, bearing) values($1, $2, ST_GeomFromWKB($3), ST_GeomFromWKB($4), $5)"
	log.Debugf("insert query: %s", query)
	result, err := conn.Exec(context.Background(), query,
		m.Timestamp,
		m.Station,
		wkb.Value(orb.Point{m.Longitude, m.Latitude}),
		wkb.Value(orb.LineString{orb.Point{m.Longitude, m.Latitude}, orb.Point{endPoint.Lng(), endPoint.Lat()}}),
		m.Bearing,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() != 1 {
		return errors.New(fmt.Sprintf("insert resulted in %d amount of rows, instead of 1", result.RowsAffected()))
	}
	return nil
}

func (d *TimescaleDB) GetPositions(since time.Duration) (positions []*types.Position, err error) {
	if since.Seconds() < 1 {
		return nil, errors.New("since should be >= 1")
	}
	if d.connectionPool == nil {
		return nil, errors.New("please connect to the database first")
	}
	conn, err := d.connectionPool.Acquire(context.Background())
	if err != nil {
		log.Errorf("failed to get connection from database pool: %e", err)
		return nil, err
	}

	defer conn.Release()

	// get average / center point
	query := fmt.Sprintf("select time, station, ST_AsBinary(point) from doppler where time > NOW() - interval '%d seconds'", int(since.Seconds()))
	log.Debugf("get positions query: %s", query)
	rows, err := conn.Query(context.Background(), query)

	if err != nil {
		log.Errorf("failed to run query: %e", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			datetime time.Time
			station  string
			point    orb.Point
		)

		err := rows.Scan(&datetime, &station, wkb.Scanner(&point))
		if err != nil {
			log.Errorf("failed to get row: %e", err)
			return nil, err
		}

		position := types.Position{
			Timestamp: datetime,
			Station:   station,
			Longitude: point.X(),
			Latitude:  point.Y(),
		}
		positions = append(positions, &position)
		log.Debugf("got position: %v", position)
	}
	err = rows.Err()
	return
}

func (d *TimescaleDB) GetLines(since time.Duration) (lines []*types.Line, err error) {
	if since.Seconds() < 1 {
		return nil, errors.New("since should be >= 1")
	}
	if d.connectionPool == nil {
		return nil, errors.New("please connect to the database first")
	}
	conn, err := d.connectionPool.Acquire(context.Background())
	if err != nil {
		log.Errorf("failed to get connection from database pool: %e", err)
		return nil, err
	}

	defer conn.Release()

	query := fmt.Sprintf("select time, station, ST_AsBinary(line) from doppler where time > NOW() - interval '%d seconds'", int(since.Seconds()))
	log.Debugf("get lines query: %s", query)
	rows, err := conn.Query(context.Background(), query)

	if err != nil {
		log.Errorf("failed to run query: %e", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			datetime time.Time
			station  string
			line     orb.LineString
		)

		err := rows.Scan(&datetime, &station, wkb.Scanner(&line))
		if err != nil {
			log.Errorf("failed to get row: %e", err)
			return nil, err
		}
		newLine := types.Line{
			Position: types.Position{
				Timestamp: datetime,
				Station:   station,
				Longitude: line[0].X(),
				Latitude:  line[0].Y(),
			},
			LongitudeEnd: line[1].X(),
			LatitudeEnd:  line[1].Y(),
		}
		lines = append(lines, &newLine)
		log.Debugf("got line: %v", newLine)
	}
	err = rows.Err()
	return
}

func (d *TimescaleDB) GetCrossings(since time.Duration) (crossings []*types.Crossing, err error) {
	if since.Seconds() < 1 {
		return nil, errors.New("since should be >= 1")
	}
	if d.connectionPool == nil {
		return nil, errors.New("please connect to the database first")
	}
	conn, err := d.connectionPool.Acquire(context.Background())
	if err != nil {
		log.Errorf("failed to get connection from database pool: %e", err)
		return nil, err
	}

	defer conn.Release()

	// INTERSECTION:
	query := fmt.Sprintf("select ST_AsBinary(ST_Intersection(a.line, b.line)), COUNT(a.station) FROM doppler AS a, doppler AS b WHERE ST_Intersects(a.line, b.line) AND a.station < b.station AND a.time > NOW() - interval '%d seconds' AND b.time > NOW() - interval '%d seconds' GROUP BY ST_Intersection(a.line, b.line);", int(since.Seconds()), int(since.Seconds()))
	log.Debugf("get lines query: %s", query)
	rows, err := conn.Query(context.Background(), query)

	if err != nil {
		log.Errorf("failed to run query: %e", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			crossing     orb.Point
			weight       int
		)

		err := rows.Scan(wkb.Scanner(&crossing), &weight)
		if err != nil {
			log.Errorf("failed to get row: %e", err)
			return nil, err
		}
		newCrossing := types.Crossing{
			Longitude: crossing.X(),
			Latitude: crossing.Y(),
			Weight:weight,
		}
		crossings = append(crossings, &newCrossing)
		log.Debugf("got line: %v", newCrossing)
	}
	err = rows.Err()
	return
}
