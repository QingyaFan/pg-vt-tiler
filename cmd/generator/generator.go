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

// tileToExtent tile xyz to extent array
func tileToExtent(z, x, y int) [4]float64 {
	tileWidth := half * 2 / math.Pow(2, float64(z))

	// calculate x lower and high coord
	xMin := float64(x)*tileWidth - half
	xMax := float64(x+1)*tileWidth - half

	// calculate y lower and high coord
	yMin := half - float64(y+1)*tileWidth
	yMax := half - float64(y)*tileWidth

	return [4]float64{xMin, yMin, xMax, yMax}
}

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

// getDataExtent 获取数据的空间范围
func getDataExtent(tableName string, geom string) [4]float64 {
	var extent string
	sql := fmt.Sprintf(`select st_extent(%s) as extent from %s`, geom, tableName)
	err = db.QueryRow(sql).Scan(&extent)
	if err != nil {
		log.Fatal(err)
	}

	return boxToArray(extent)
}

// GenerateTile 生成切片
func GenerateTile() {
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
}
