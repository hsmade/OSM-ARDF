#!/bin/bash
set -e

until PGPASSWORD=$POSTGRES_PASSWORD psql -h "timescaledb" -U "postgres" -c '\q'; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

>&2 echo "Postgres is up - creating DB"

PGPASSWORD=$POSTGRES_PASSWORD psql -h "timescaledb" -U "postgres" -c "create table doppler (time TIMESTAMP not null default now(), name text not null, lat real not null, lon real not null, bearing INT);"
PGPASSWORD=$POSTGRES_PASSWORD psql -h "timescaledb" -U "postgres" -c "SELECT create_hypertable('doppler', 'time', chunk_time_interval => interval '1 minute');"
