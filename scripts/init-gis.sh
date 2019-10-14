#!/bin/bash
set -xe

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE TABLE doppler (time TIMESTAMP NOT NULL DEFAULT now(), name TEXT NOT NULL, lat REAL NOT NULL, lon REAL NOT NULL, bearing INT);
    SELECT create_hypertable('doppler', 'time', chunk_time_interval => INTERVAL '1 minute');
EOSQL