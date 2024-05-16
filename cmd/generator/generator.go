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

	_ "github.com/lib/pq"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type generateConfig struct {
	DSN            string
	StartZoomLevel int
	EndZoomLevel   int
	MaxOpenConns   int
	TableName      string
	GeometryName   string
	Concurrency    int
	TileLocation   string
}

const (
	tileSize  = 256
	half      = 20037508.342789244
	shortDesc = "pg-vt-tiler is a tool that generate vector tiles, suppose original data from postgresql/postgis"
)

var (
	cfg     generateConfig
	db      *sql.DB
	err     interface{}
	cfgFile string
)

var generatorCmd = &cobra.Command{
	Use:   `pg-vt-tiler`,
	Short: shortDesc,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: check dsn string format
		// TODO: assurance start <= end
		if cfg.DSN == "" {
			log.Fatalln(`dsn should not be empty.`)
		}
		if cfg.TableName == "" {
			log.Fatalln("table must be set.")
		}
		if cfg.GeometryName == "" {
			log.Fatalln("geom column name in table must be set.")
		}
		err := Generate(cfg.TableName, cfg.GeometryName)
		if err != nil {
			fmt.Printf("Error generate tiles, err: %v \n", err)
		}
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		initDB(cfg.DSN)
	},
}

func initDB(dsn string) {
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf(`database open err: %v \n`, err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatalf(`database ping err: %v \n`, err)
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(50)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			log.Fatalf(`home-dir error: %v \n`, err)
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(".cobra")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func init() {
	dsnHint := `database connection info, format: "host=localhost port=5432 user=postgres password=postgres dbname=db_name sslmode=ssl_mode", required.`

	cobra.OnInitialize(initConfig)
	generatorCmd.PersistentFlags().StringVarP(&cfg.DSN, "dsn", "d", "", dsnHint)
	generatorCmd.PersistentFlags().IntVarP(&cfg.StartZoomLevel, "start", "s", 7, "")
	generatorCmd.PersistentFlags().IntVarP(&cfg.EndZoomLevel, "end", "e", 7, "")
	generatorCmd.PersistentFlags().StringVarP(&cfg.TableName, "table", "t", "", "")
	generatorCmd.PersistentFlags().StringVarP(&cfg.GeometryName, "geom", "g", "", "")
	generatorCmd.PersistentFlags().StringVarP(&cfg.TileLocation, "location", "l", ".", "")
	generatorCmd.PersistentFlags().IntVarP(&cfg.Concurrency, "concurrency", "c", 10, "")
	generatorCmd.PersistentFlags().IntVarP(&cfg.MaxOpenConns, "dbconns", "n", 50, "")

	requiredFlags := []string{"dsn", "start", "end", "table", "geom"}
	for _, flag := range requiredFlags {
		err := generatorCmd.MarkPersistentFlagRequired(flag)
		if err != nil {
			log.Fatalf("Error marking flag %s as required, err: %v \n", flag, err)
		}
	}
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

	tileName := fmt.Sprintf(`%v/%d/%d/%d.pbf`, cfg.TileLocation, z, x, y)
	err = ioutil.WriteFile(tileName, mvt, 0755)
	if err != nil {
		log.Fatal(err)
	}
}

func createZoomLevelDirectoryStructure(zoomLevel int, startX, endX int) {
	for i := startX; i <= endX; i++ {
		pathName := fmt.Sprintf(`%v/%v/%v`, cfg.TileLocation, zoomLevel, i)
		os.MkdirAll(pathName, 0755)
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
	for z := cfg.StartZoomLevel; z <= cfg.EndZoomLevel; z++ {
		interval := half * 2 / math.Pow(2, float64(z))
		xMin := int(math.Floor((dataExtent[0] + half) / interval))
		yMin := int(math.Floor((half - dataExtent[3]) / interval))
		xMax := int(math.Floor((dataExtent[2] + half) / interval))
		yMax := int(math.Floor((half - dataExtent[1]) / interval))
		createZoomLevelDirectoryStructure(z, xMin, xMax)
		for x := xMin; x <= xMax; x++ {
			for y := yMin; y <= yMax; y++ {
				zxyRange = append(zxyRange, [3]int{z, x, y})
				total++
			}
		}
	}

	// 对于每个切片，转换为坐标范围，并生成切片，写入文件
	waitCh := make(chan [3]int, cfg.Concurrency)
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

// Execute cobra的入口函数
func Execute() {
	err := generatorCmd.Execute()
	if err != nil {
		log.Fatalf("generate tiles failed, err: %v \n", err)
	}
}
