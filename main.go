package main

import (
	"database/sql"
	"fmt"
	"github.com/paulmach/orb"
	"log"
	"os"
	"strconv"

	"github.com/fogleman/gg"
	"github.com/lib/pq" // Import for postgres
	"github.com/paulmach/orb/encoding/wkb"
	"github.com/paulmach/orb/geojson"
)

type MinMax struct {
	latMin float64 // = 90.0
	latMax float64 // = -90.0
	lonMin float64 // = 180.0
	lonMax float64 // = -180.0
}

func main() {
	LOCAL, _ := strconv.ParseBool(os.Getenv("LOCAL"))
	const URLPrefix string = "https://maptiles.glidercheck.com/hypsometric"
	for i := 0; i <= 1; i++ {
		for j := 0; j <= 1; j++ {
			downloadFile(fmt.Sprintf("image_%s_%s.jpeg", fmt.Sprint(i), fmt.Sprint(j)), fmt.Sprintf("%s/1/%s/%s.jpeg", URLPrefix, fmt.Sprint(i), fmt.Sprint(j)))
		}
	}

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

func FindMinMax(line orb.LineString) MinMax {
	minmax := MinMax{90.0, -90.0, 180.0, -180.0}
	for _, p := range line {
		if p.Lat() < minmax.latMin {
			minmax.latMin = p.Lat()
		}
		if p.Lat() > minmax.latMax {
			minmax.latMax = p.Lat()
		}
		if p.Lon() < minmax.lonMin {
			minmax.lonMin = p.Lon()
		}
		if p.Lon() > minmax.lonMax {
			minmax.lonMax = p.Lon()
		}
	}
	return minmax
}

func Normalize(line orb.LineString, minmax MinMax) orb.LineString {
	for i := range line {
		// p[1] == p.Lat()
		line[i][1] = (line[i][1] - minmax.latMin) / (minmax.latMax - minmax.latMin)
		// p[0] == p.Lon()
		line[i][0] = (line[i][0] - minmax.lonMin) / (minmax.lonMax - minmax.lonMin)
	}
	return line
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
	rows, err := db.Query("SELECT ST_AsBinary(line_wkt),bbox from flight where id='10'")

	// MergeImage()
	MergeImage4_4()

	for rows.Next() {
		// Array for postgres query
		arr := pq.Float64Array{}

		// parse to ST_AsBinary(line_wkt) and bbox to arr
		err := rows.Scan(wkb.Scanner(&line), &arr)

		// Cast postgres array to native go array
		bbox := []float64(arr)
		log.Println(bbox)

		feature := geojson.NewFeature(line)

		// Convert to lineString (the syntax Geometry. is necessary due to the interface)
		line := feature.Geometry.(orb.LineString)
		// open image
		im, err := gg.LoadJPG("images/map.jpg")
		if err != nil {
			panic(err)
		}
		// pattern := gg.NewSurfacePattern(im, gg.RepeatBoth)
		dc := gg.NewContextForImage(im)
		// dc := gg.NewContext(1024, 1024)
		log.Println(dc.Height(), dc.Width())
		dc.SetRGB(1, 1, 1)
		// minmax := FindMinMax(line)
		log.Println(FindMinMax(line))
		minmax := MinMax{40.97, 55.77, 0, 22.5}
		line = Normalize(line, minmax)
		dc.SetRGB(0.175, 0.33, 0.65)
		for _, p := range line {
			// the origin of the canvas is located at the top left corner
			// therefore the coordinates have to be rotated
			// https://en.wikipedia.org/wiki/Rotation_matrix
			// Plot		  x				,  y
			dc.DrawCircle(p.Lon()*512+10, (1-p.Lat())*512, 1.0)
			dc.Fill()
		}
		dc.Fill()
		dc.SavePNG("images/out.png")
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return nil, nil
}
