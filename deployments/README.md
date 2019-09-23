# Example deployment
This docker-compose will start the following containers:

 * mapcache: This is the caching proxy for OSM
 * timescaledb: This is the database that will contain the measurements
 * timescaledb_init: Creates the DB table
 * aprs_receiver: receives APRS broadcasts
 * udp_receiver: receives UDP measurements
 * webserver: serves the API that generates the layers on OSM