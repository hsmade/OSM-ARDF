package database

import "github.com/hsmade/OSM-ARDF/pkg/measurement"

type Database interface {
	Connect() error
	Add(m *measurement.Measurement) error
}
