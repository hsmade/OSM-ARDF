package stdin

import (
	"bytes"
	"errors"
	"github.com/hsmade/OSM-ARDF/pkg/measurement"
	"reflect"
	"testing"
	"time"
)

type databaseMock struct {
	value        *measurement.Measurement
	measurements int
	Test         *testing.T
}

func (d *databaseMock) Add(m *measurement.Measurement) error {
	d.Test.Logf("Add(%v)", *m)
	d.value = m
	d.measurements++
	return nil
}

func (d *databaseMock) Connect() error {
	return nil
}

type databaseMockNoConnect struct {
	value        *measurement.Measurement
	measurements int
	Test         *testing.T
}

func (d *databaseMockNoConnect) Add(m *measurement.Measurement) error {
	d.Test.Logf("Add(%v)", *m)
	d.value = m
	d.measurements++
	return nil
}

func (d *databaseMockNoConnect) Connect() error {
	return errors.New("test")
}

func TestReceiver_Start_Happy_flow(t *testing.T) {
	db := &databaseMock{Test: t}
	r := &Receiver{Database: db}
	err := r.Start(bytes.NewReader([]byte("{}")))
	if err != nil {
		t.Errorf("Got unexpected error %v", err)
	}

	if !reflect.DeepEqual(measurement.Measurement{}, *db.value) {
		t.Errorf("Got unexpected measurement, should be empty: %v", *db.value)
	}

	if db.measurements != 1 {
		t.Errorf("Got unexpected amount of measurements (need 1): %v", db.measurements)
	}
}

func TestReceiver_Start_No_connect(t *testing.T) {
	db := &databaseMockNoConnect{Test: t}
	r := &Receiver{Database: db}
	err := r.Start(bytes.NewReader([]byte("{}")))
	if !reflect.DeepEqual(err, errors.New("test")) {
		t.Errorf("Got unexpected error %v", err)
	}

	if db.value != nil {
		t.Errorf("Got unexpected measurement, should be empty: %v", *db.value)
	}

	if db.measurements != 0 {
		t.Errorf("Got unexpected amount of measurements (need 0): %v", db.measurements)
	}
}

func TestReceiver_process(t *testing.T) {
	validTIme := time.Time{}
	_ = validTIme.UnmarshalText([]byte("2018-09-22T12:42:31Z"))

	tests := []struct {
		name   string
		data   string
		result measurement.Measurement
	}{
		{
			name: "happy flow",
			data: "{\"timestamp\":\"2018-09-22T12:42:31Z\", \"station\":\"abc\", \"longitude\": 52.5, \"latitude\": 5.0, \"bearing\": 180}",
			result: measurement.Measurement{
				Timestamp: validTIme,
				Station:   "abc",
				Longitude: 52.5,
				Latitude:  5.0,
				Bearing:   180,
			},
		},
		{
			name:   "broken time",
			data:   "{\"timestamp\":\"garbage\", \"station\":\"abc\", \"longitude\": 52.5, \"latitude\": 5.0, \"bearing\": 180}",
			result: measurement.Measurement{},
		},
		{
			name:   "invalid json",
			data:   "garbage",
			result: measurement.Measurement{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &databaseMock{Test: t, value: &measurement.Measurement{}}
			r := &Receiver{Database: db}
			r.process(tt.data)
			if !reflect.DeepEqual(*db.value, tt.result) {
				t.Errorf("Process() = %v, want %v", *db.value, tt.result)
			}
		})
	}
}
