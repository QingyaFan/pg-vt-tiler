#!/bin/bash

./pg-vt-tiler \
    -d="host=127.0.0.1 port=5432 user=postgres dbname=postgres sslmode=disable" \
    -s=5 -e=5 \
    -t="public.planet_osm_polygon" -g=way \
    -l=data