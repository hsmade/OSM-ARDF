package datastructures

import "time"

type Position struct {
	Timestamp time.Time
	Station   string
	Longitude float64
	Latitude  float64
}
