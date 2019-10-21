#!/bin/bash
set -e

until PGPASSWORD=$POSTGRES_PASSWORD psql -h "timescaledb" -U "postgres" -c '\q'; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

>&2 echo "Postgres is up - creating DB"

/init-git.sh