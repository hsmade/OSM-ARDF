FROM timescale/timescaledb-postgis:1.4.2-pg11
ADD init.sh /init.sh
ADD init-gis.sh /init-gis.sh
ENTRYPOINT ["/init.sh"]