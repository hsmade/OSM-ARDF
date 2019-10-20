package types

import (
	"time"
)

type Measurement struct {
	Timestamp time.Time
	Station   string
	Longitude float64
	Latitude  float64
	Bearing   int
}
