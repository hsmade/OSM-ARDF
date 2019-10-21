#!/bin/bash
set -xe
#    CREATE TABLE doppler (time TIMESTAMPTZ NOT NULL DEFAULT now(), name TEXT NOT NULL, lat REAL NOT NULL, lon REAL NOT NULL, bearing INT);

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE EXTENSION postgis;
    CREATE TABLE doppler (time TIMESTAMPTZ NOT NULL DEFAULT now(), station TEXT NOT NULL, point GEOMETRY, bearing INT);
    SELECT create_hypertable('doppler', 'time', chunk_time_interval => INTERVAL '1 minute');
EOSQL