#!/bin/bash

osm2pgsql \
   --create \
   --database postgres \
   --user postgres \
   --host 127.0.0.1 \
   --port 5432 \
   --schema public \
   data/xxx-latest.osm.pbf
