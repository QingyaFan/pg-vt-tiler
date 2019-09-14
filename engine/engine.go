package engine

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "cheerfun"
	password = "cheerfun"
	dbname   = "osm"
)

func getTileBoundryByZoomLevel(zoom int) map[byte]int {

	// get extent of data
	sql := `select st_extent(geom) as extent from water`
	
}

// GenerateTile generate tile
func GenerateTile() {
	connInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", connInfo)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	var mvt []byte

	sql := `select st_asmvt(r) as mvt 
		from
		(select st_asmvtgeom(a.geom, a.extent) as mvtgeom
		from (select st_extent(geom) as extent, geom from water where gid = 1 group by geom) a)
		as r;`

	err = db.QueryRow(sql).Scan(&mvt)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("test.pbf", mvt, 0755)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(mvt[:]))
}
