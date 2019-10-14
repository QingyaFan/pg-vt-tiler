package generator

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	// pq
	_ "github.com/lib/pq"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	tileSize = 256
	half     = 20037508.342789244
)

var (
	db             *sql.DB
	err            interface{}
	cfgFile        string
	dsn            string
	tableName      string
	geometryName   string
	startZoomLevel int
	endZoomLevel   int
	generatorCmd   = &cobra.Command{
		Use: `pg-vt-tiler`,
		Short: `pg-vt-tiler 是一个生成矢量瓦片数据集的工具，
数据源是PostgreSQL中存储的矢量数据。`,
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: 检查dsn格式
			// TODO: 检查start和end的范围，并保证start<=end
			if dsn == "" {
				fmt.Println(`dsn必须设定`)
				os.Exit(1)
			}
			if tableName == "" || geometryName == "" {
				fmt.Println(`table和geom必须设定`)
				os.Exit(1)
			}
			Generate(`water`, `geom`)
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			initDB(dsn)
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	generatorCmd.PersistentFlags().StringVarP(&dsn, "dsn", "d", "", `database connection info, format: "host=localhost port=5432 user=postgres password=postgres dbname=db_name sslmode=ssl_mode", required.`)
	generatorCmd.PersistentFlags().IntVarP(&startZoomLevel, "start", "s", 7, "")
	generatorCmd.PersistentFlags().IntVarP(&endZoomLevel, "end", "e", 7, "")
	generatorCmd.PersistentFlags().StringVarP(&tableName, "table", "t", "", "")
	generatorCmd.PersistentFlags().StringVarP(&geometryName, "geom", "g", "", "")
	generatorCmd.MarkFlagRequired("dsn")
	generatorCmd.MarkFlagRequired("start")
	generatorCmd.MarkFlagRequired("end")
	generatorCmd.MarkFlagRequired("table")
	generatorCmd.MarkFlagRequired("geom")
}

func initDB(dsn string) {
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(`database open error`, err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(`database ping error`, err)
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
	err = db.QueryRow(sql).Scan(&mvt)
	if err != nil {
		log.Fatal(err)
	}

	tileName := fmt.Sprintf(`/Volumes/Samsung_T5/tmp/%d.%d.%d.pbf`, z, x, y)
	err = ioutil.WriteFile(tileName, mvt, 0755)
	if err != nil {
		log.Fatal(err)
	}
}

// Generate 生成指定数据指定范围的矢量切片
func Generate(tableName string, geom string) error {
	if tableName == "" || geom == "" {
		log.Print(`数据表名和空间字段都不能为空`)
		return errors.New(`数据表名和空间字段都不能为空`)
	}
	// 获取数据范围，计算xy的范围
	dataExtent := getDataExtent(tableName, geom)
	var zxyRange [][3]int
	total := 0
	for z := startZoomLevel; z <= endZoomLevel; z++ {
		interval := half * 2 / math.Pow(2, float64(z))
		xMin := int(math.Floor((dataExtent[0] + half) / interval))
		yMin := int(math.Floor((half - dataExtent[3]) / interval))
		xMax := int(math.Floor((dataExtent[2] + half) / interval))
		yMax := int(math.Floor((half - dataExtent[1]) / interval))
		for x := xMin; x <= xMax; x++ {
			for y := yMin; y <= yMax; y++ {
				zxyRange = append(zxyRange, [3]int{z, x, y})
				total++
			}
		}
	}

	// 对于每个切片，转换为坐标范围，并生成切片，写入文件
	waitCh := make(chan [3]int)
	for _, tile := range zxyRange {
		go func(t [3]int) {
			generateTile(t[0], t[1], t[2], tableName, geom)
			waitCh <- t
		}(tile)
	}
	tmp := total
	for range zxyRange {
		<-waitCh
		tmp--
		fmt.Printf("完成: %v/%v\n", total-tmp, total)
	}

	return nil
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(`home-dir error`, err)
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(".cobra")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// Execute cobra的入口函数
func Execute() {
	if err := generatorCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
