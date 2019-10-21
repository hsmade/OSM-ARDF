    CREATE EXTENSION IF NOT EXISTS postgis;
    CREATE TABLE doppler (time TIMESTAMPTZ NOT NULL DEFAULT now(), station TEXT NOT NULL, point GEOMETRY, line GEOMETRY, bearing INT);
    SELECT create_hypertable('doppler', 'time', chunk_time_interval => INTERVAL '1 minute');
