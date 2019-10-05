package database

import (
	"github.com/hsmade/OSM-ARDF/pkg/measurement"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestTimescaleDB_Add(t *testing.T) {
	dockerPort, err := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	if err != nil {
		t.Fatalf("failed to get port for postgresql instance running in docker: %e", err)
	}

	dockerIp := os.Getenv("POSTGRES_IP")
	if dockerIp == "" {
		t.Fatalf("failed to get ip for postgresql instance running in docker: %e", err)
	}

	testMeasurement := measurement.Measurement{
		Timestamp: time.Now(),
		Station:   "test",
		Longitude: 1,
		Latitude:  2,
		Bearing:   3,
	}

	type fields struct {
		Host         string
		Port         uint16
		Username     string
		Password     string
		DatabaseName string
	}
	type args struct {
		m *measurement.Measurement
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Happy path",
			fields{
				Host:         dockerIp,
				Port:         uint16(dockerPort),
				Username:     "postgres",
				Password:     "postgres",
				DatabaseName: "postgres",
			},
			args{
				&testMeasurement,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &TimescaleDB{
				Host:     tt.fields.Host,
				Port:     tt.fields.Port,
				Username: tt.fields.Username,
				Password: tt.fields.Password,
			}
			err := d.Connect()
			if err != nil {
				t.Errorf("Failed during Connect(): %e", err)
			}
			if err := d.Add(tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTimescaleDB_Connect(t *testing.T) {
	dockerPort, err := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	if err != nil {
		t.Fatalf("failed to get port for postgresql instance running in docker: %e", err)
	}
	dockerIp := os.Getenv("POSTGRES_IP")
	if dockerIp == "" {
		t.Fatalf("failed to get ip for postgresql instance running in docker: %e", err)
	}
	type fields struct {
		Host         string
		Port         uint16
		Username     string
		Password     string
		DatabaseName string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Happy path",
			fields: fields{
				Host:         dockerIp,
				Port:         uint16(dockerPort),
				Username:     "postgres",
				Password:     "postgres",
				DatabaseName: "postgres",
			},
			wantErr: false,
		},
		{
			name: "Connection failed",
			fields: fields{
				Host:     "900.900.900.900",
				Port:     0,
				Username: "postgres",
				Password: "postgres",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &TimescaleDB{
				Host:     tt.fields.Host,
				Port:     tt.fields.Port,
				Username: tt.fields.Username,
				Password: tt.fields.Password,
			}
			if err := d.Connect(); (err != nil) != tt.wantErr {
				t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
