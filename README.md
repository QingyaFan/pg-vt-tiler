# PG-VT-TILER

Generate vector tiles from data stored in postgresql/postgis.

It is mainly focus on generating vector tiles using the vector tile functions in postgis. For example:

- ST_AsMVTGeom
- ST_AsMVT

It read geometry column from specified table, and generate vector tile, and store the result to pg. However, it can write the results to other storage, like redis, seaweed and so on. So the results storage may be a plugin mechanism.
