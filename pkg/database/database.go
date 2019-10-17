package database

import (
	"github.com/hsmade/OSM-ARDF/pkg/types"
	"time"
)

type Database interface {
	Connect() error
	Add(m *types.Measurement) error
	GetPositions(since time.Duration) ([]*types.Position, error)
}
