package main

import (
	"pg-vt-tiler/engine"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "cheerfun"
	dbname   = "g-default"
)

// accept table name and geometry column name
func main() {
	engine.GenerateTile()
}
