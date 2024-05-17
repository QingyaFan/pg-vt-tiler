#!/bin/bash

docker run \
  --name postgis \
  --user postgres \
  --platform linux/amd64 \
  --env POSTGRES_PASSWORD=postgis \
  --env POSTGRES_HOST_AUTH_METHOD=trust \
  --publish 5432:5432 \
  --detach \
  postgis/postgis:16-3.4