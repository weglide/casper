package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

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
		// TODO: check if pbf is really necessary?
		// e.g. localhost:7979/flights/12,13/
		test_line_wkt()
		// r.HandleFunc("/flights/{ids}", flightHandlerLocal)
		// http.ListenAndServe(":7979", r)
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
func test_line_wkt() (*geojson.FeatureCollection, error) {

	// open connection
	db, err := sql.Open("postgres", psqlConnectionString())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// execute query
	rows, err := db.Query("SELECT takeoff_airport_id from flight where id='11'")
	log.Printf("Rows: %d", rows)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var id int
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(10)
		log.Println(id)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	// // var geo geom.Geometry
	// geo, err = DecodeBytes(rows)
	// if err != nil {
	// 	return nil, err
	// }
	// defer rows.Close()

	featureCollection := geojson.NewFeatureCollection()
	return featureCollection, nil
}
