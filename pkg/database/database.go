package database

import (
	"github.com/hsmade/OSM-ARDF/pkg/datastructures"
	"time"
)

type Database interface {
	Connect() error
	Add(m *datastructures.Measurement) error
	GetPositions(since time.Duration) ([]*datastructures.Position, error)
}
