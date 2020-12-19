package main

import (
	"database/sql"
	"fmt"
	"github.com/paulmach/orb"
	"log"
	"os"
	"strconv"

	"github.com/fogleman/gg"
	_ "github.com/lib/pq" // Import for postgres
	"github.com/paulmach/orb/encoding/wkb"
	"github.com/paulmach/orb/geojson"
)

// required input params
// type Params struct {
// 	ids  []int
// 	zoom maptile.Zoom
// 	x    uint32
// 	y    uint32
// }

// type Response struct {
// 	body            string
// 	isBase64Encoded bool
// 	statusCode      int
// 	headers         string
// }

// type Tile struct {
// 	Z    int
// 	X    int
// 	Y    int
// 	Lat  float64
// 	Long float64
// }

// type Conversion interface {
// 	deg2num(t *Tile) (x int, y int)
// 	num2deg(t *Tile) (lat float64, long float64)
// }

func main() {
	LOCAL, _ := strconv.ParseBool(os.Getenv("LOCAL"))

	// switch between lambda and local environment
	if LOCAL == true {
		// r := mux.NewRouter()
		//								ids  /z/x/y
		// e.g. localhost:7979/flights/12,13/
		test_line_wkt()

	}
	// else {
	//lambda.Start(flightHandlerLambda)
	// }
}

func psqlConnectionString() string {
	// get environment connection vars
	var (
		host     = os.Getenv("POSTGRES_HOST")
		port     = os.Getenv("POSTGRES_PORT")
		user     = os.Getenv("POSTGRES_USER")
		password = os.Getenv("POSTGRES_PASS")
		dbname   = os.Getenv("POSTGRES_DB")
	)

	// build connection string
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
}

// fetch line strings from db by ids
func test_line_wkt() (error, error) {

	var line orb.LineString
	// open connection
	log.Println(psqlConnectionString())
	db, err := sql.Open("postgres", psqlConnectionString())
	if err != nil {
		log.Println("DB Error connection failed")
		return nil, err
	}
	defer db.Close()

	// execute query
	rows, err := db.Query("SELECT ST_AsBinary(line_wkt) from flight where id='11'")

	for rows.Next() {

		err := rows.Scan(wkb.Scanner(&line))
		if err != nil {
			return nil, err
		}
		feature := geojson.NewFeature(line)
		feature.Properties = geojson.Properties{
			"id": 11,
		}
		if err != nil {
			log.Fatal(err)
		}

		// Convert to lineString
		line := feature.Geometry.(orb.LineString)
		log.Println(line[0])
		dc := gg.NewContext(1024, 1024)
		dc.SetRGB(1, 1, 1)
		//dc.Fill()
		// for _, p := range line {
		// 	dc.DrawCircle(p[0], p[1], 1)
		// 	log.Println(p[0], p[1])
		// 	dc.Fill()
		// }
		dc.SetRGBA(0, 0, 1, 1)
		dc.DrawCircle(0.75, 0, 40)
		dc.Fill()
		dc.SetRGBA(1, 1, 1, 1)
		dc.DrawCircle(100, 0.5, 10)
		dc.Fill()
		// dc.DrawCircle(500, 500, 400)
		// dc.SetRGB(0, 0, 0)
		dc.SavePNG("out.png")
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return nil, nil
}
