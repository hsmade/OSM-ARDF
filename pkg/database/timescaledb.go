package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/apex/log"
	"github.com/hsmade/OSM-ARDF/pkg/datastructures"
	"github.com/jackc/pgx/v4/pgxpool"
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

func (d *TimescaleDB) Add(m *datastructures.Measurement) error {
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

	query := "insert into \"doppler\"(time, name, lon, lat, bearing) values($1, $2, $3, $4, $5)"
	log.Debugf("insert query: %s", query)
	result, err := conn.Exec(context.Background(), query,
		m.Timestamp,
		m.Station,
		m.Longitude,
		m.Latitude,
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

func (d *TimescaleDB) GetPositions(since time.Duration) (positions []*datastructures.Position, err error) {
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

	query := fmt.Sprintf("select date_trunc('second', time) as second, name, avg(lon), avg(lat) from doppler where time > NOW() - interval '%d seconds' group by second,name order by second, name", int(since.Seconds()))
	log.Debugf("get positions query: %s", query)
	rows, err := conn.Query(context.Background(), query)

	if err != nil {
		log.Errorf("failed to run query: %e", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			datetime  time.Time
			name      string
			longitude float64
			latitude  float64
		)

		err := rows.Scan(&datetime, &name, &longitude, &latitude)
		if err != nil {
			log.Errorf("failed to get row: %e", err)
			return nil, err
		}

		position := datastructures.Position{
			Timestamp: datetime,
			Station:   name,
			Longitude: longitude,
			Latitude:  latitude,
		}
		positions = append(positions, &position)
		log.Debugf("got position: %v", position)
	}
	err = rows.Err()
	return
}
