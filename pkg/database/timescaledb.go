package database

import (
	"context"
	"fmt"
	"github.com/hsmade/OSM-ARDF/pkg/measurement"
	"github.com/jackc/pgx/v4/pgxpool"
)

type TimescaleDB struct {
	Host         string
	Port         uint16
	Username     string
	Password     string
	DatabaseName string
	connection   *pgxpool.Pool
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
	d.connection = pool
	return err
}

func (d *TimescaleDB) Add(m *measurement.Measurement) error {

	return nil
}
