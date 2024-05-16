# PG-VT-TILER

PG-VT-Tiller can help generating vector tiles from geographic data stored in postgresql/postgis. Simply put, it use the vector tile functions in postgis:  ST_AsMVTGeom and ST_AsMVT.

It can read geometry column from specified table, and generate vector tile, and store the result to a file directory. However, it can write the results to other storage, like redis, seaweed and so on. So the results storage may be a plugin mechanism.(TODO:)

## Usage

Download the binary file from [release page](https://github.com/QingyaFan/pg-vt-tiler/releases).

You can see it's options by `tiler -h`:

```txt
Usage:
  pg-vt-tiler [flags]

Flags:
  -c, --concurrency int    (default 10)
  -d, --dsn string        database connection info, format: "host=localhost port=5432 user=postgres password=postgres dbname=db_name sslmode=ssl_mode", required.
  -t, --table string
  -g, --geom string
  -l, --location string    (default ".")
  -s, --start int          (default 7)
  -e, --end int            (default 7)
  -h, --help              help for pg-vt-tiler
```

## Linux

In linux, you can move the binary to `/usr/local/bin`: `mv tiler /usr/local/bin`, then you can use it directly by: `tiler [options]`.
