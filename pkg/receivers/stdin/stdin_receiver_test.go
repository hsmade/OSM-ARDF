package stdin

import (
	"bytes"
	"errors"
	"github.com/hsmade/OSM-ARDF/pkg/types"
	"reflect"
	"testing"
	"time"
)

type databaseMock struct {
	value        *types.Measurement
	measurements int
	Test         *testing.T
}

func (d *databaseMock) Add(m *types.Measurement) error {
	d.Test.Logf("Add(%v)", *m)
	d.value = m
	d.measurements++
	return nil
}

func (d *databaseMock) Connect() error {
	return nil
}

func (d *databaseMock) GetPositions(since time.Duration) ([]*types.Position, error) {
	return nil, nil
}

func (d *databaseMock) GetLines(since time.Duration) ([]*types.Line, error) {
	return nil, nil
}

func (d *databaseMock) GetCrossings(since time.Duration) ([]*types.Crossing, error) {
	return nil, nil
}

type databaseMockNoConnect struct {
	value        *types.Measurement
	measurements int
	Test         *testing.T
}

func (d *databaseMockNoConnect) Add(m *types.Measurement) error {
	d.Test.Logf("Add(%v)", *m)
	d.value = m
	d.measurements++
	return nil
}

func (d *databaseMockNoConnect) Connect() error {
	return errors.New("test")
}

func (d *databaseMockNoConnect) GetPositions(since time.Duration) ([]*types.Position, error) {
	return nil, nil
}

func (d *databaseMockNoConnect) GetLines(since time.Duration) ([]*types.Line, error) {
	return nil, nil
}

func (d *databaseMockNoConnect) GetCrossings(since time.Duration) ([]*types.Crossing, error) {
	return nil, nil
}

func TestReceiver_Start_Happy_flow(t *testing.T) {
	db := &databaseMock{Test: t}
	r := &Receiver{Database: db}
	err := r.Start(bytes.NewReader([]byte("{}")))
	if err != nil {
		t.Errorf("Got unexpected error %v", err)
	}

	if !reflect.DeepEqual(types.Measurement{}, *db.value) {
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
		result types.Measurement
	}{
		{
			name: "happy flow",
			data: "{\"timestamp\":\"2018-09-22T12:42:31Z\", \"station\":\"abc\", \"longitude\": 52.5, \"latitude\": 5.0, \"bearing\": 180}",
			result: types.Measurement{
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
			result: types.Measurement{},
		},
		{
			name:   "invalid json",
			data:   "garbage",
			result: types.Measurement{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &databaseMock{Test: t, value: &types.Measurement{}}
			r := &Receiver{Database: db}
			r.process(tt.data)
			if !reflect.DeepEqual(*db.value, tt.result) {
				t.Errorf("Process() = %v, want %v", *db.value, tt.result)
			}
		})
	}
}
