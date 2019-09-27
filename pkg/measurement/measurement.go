package measurement

import (
	"time"
)

type Measurement struct {
	Timestamp time.Time
	Station   string
	Longitude float32
	Latitude  float32
	Bearing   float32
}
