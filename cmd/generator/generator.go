package generator

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"strconv"
	"strings"

	// pq
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "cheerfun"
	password = "cheerfun"
	dbname   = "osm"
	tileSize = 256
	half     = 20037508.342789244
)

var db *sql.DB
var err interface{}

func init() {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(5)
}

// boxToArray postgis获取的box转为float64数组
func boxToArray(box string) [4]float64 {
	strArray := strings.Split(strings.ReplaceAll(strings.Trim(box, `BOX()`), ` `, `,`), `,`)
	var res [4]float64
	for i := 0; i < 4; i++ {
		res[i], err = strconv.ParseFloat(strArray[i], 64)
		if err != nil {
			log.Fatal(err)
		}
	}

	return res
}

// tileToExtent tile xyz to extent array
func tileToExtent(z, x, y int) [4]float64 {
	tileWidth := half * 2 / math.Pow(2, float64(z))

	xMin := float64(x)*tileWidth - half
	xMax := float64(x+1)*tileWidth - half

	yMin := half - float64(y+1)*tileWidth
	yMax := half - float64(y)*tileWidth

	return [4]float64{xMin, yMin, xMax, yMax}
}

// getDataExtent 获取数据整体的空间范围
func getDataExtent(tableName string, geom string) [4]float64 {
	var extent string
	sql := fmt.Sprintf(`select st_extent(%s) as extent from %s`, geom, tableName)
	err = db.QueryRow(sql).Scan(&extent)
	if err != nil {
		log.Fatal(err)
	}

	return boxToArray(extent)
}

// generateTile 生成切片
func generateTile(z, x, y int, tableName, geom string) {
	tileExtent := tileToExtent(z, x, y)
	var mvt []byte
	geomInExtentSQL := fmt.Sprintf(
		`select 
		st_intersection(%s, st_makeenvelope(%f, %f, %f, %f, %d)) as geom 
		from %s 
		where st_intersects(%s, st_makeenvelope(%f, %f, %f, %f, %d))`,
		geom,
		tileExtent[0],
		tileExtent[1],
		tileExtent[2],
		tileExtent[3],
		3857,
		tableName,
		geom,
		tileExtent[0],
		tileExtent[1],
		tileExtent[2],
		tileExtent[3],
		3857,
	)
	sql := fmt.Sprintf(`
		select st_asmvt(d) 
		from
		(
			select st_asmvtgeom(c.geom, c.extent)
			from
			(
				select st_union(b.geom) as geom, st_extent(b.geom) as extent 
				from
				(
					%s
				) as b
			) as c
		) as d`, geomInExtentSQL)
	// fmt.Println(sql)
	// return
	err = db.QueryRow(sql).Scan(&mvt)
	if err != nil {
		log.Fatal(err)
	}

	tileName := fmt.Sprintf(`%d.%d.%d.pbf`, z, x, y)
	err = ioutil.WriteFile(tileName, mvt, 0755)
	if err != nil {
		log.Fatal(err)
	}
}

func main(tableName string, geom string, zoom int) {

}
