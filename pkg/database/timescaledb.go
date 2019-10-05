package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/hsmade/OSM-ARDF/pkg/measurement"
	"github.com/jackc/pgx/v4/pgxpool"
)

type TimescaleDB struct {
	Host           string
	Port           uint16
	Username       string
	Password       string
	DatabaseName   string
	connectionPool *pgxpool.Pool
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

func (d *TimescaleDB) Add(m *measurement.Measurement) error {
	conn, err := d.connectionPool.Acquire(context.Background())
	if err != nil {
		return err
	}

	defer conn.Release()

	result, err := conn.Exec(context.Background(), "insert into doppler(time, name,lon,lat, bearing) values(%s, %s, %s, %s) ",
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
		return errors.New(fmt.Sprintf("Insert result in %d amount of rows, instead of 1", result.RowsAffected()))
	}
	return nil
}
