package database

import (
	"github.com/hsmade/OSM-ARDF/pkg/measurement"
	"testing"
)

func TestTimescaleDB_Add(t *testing.T) {
	type fields struct {
		Host     string
		Port     uint16
		Username string
		Password string
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &TimescaleDB{
				Host:     tt.fields.Host,
				Port:     tt.fields.Port,
				Username: tt.fields.Username,
				Password: tt.fields.Password,
			}
			if err := d.Add(tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTimescaleDB_Connect(t *testing.T) {
	type fields struct {
		Host     string
		Port     uint16
		Username string
		Password string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Happy path",
			fields: fields{
				Host:     "localhost",
				Port:     5432,
				Username: "postgres",
				Password: "postgres",
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
