package main

import (
	"database/sql"
	"fmt"
	"github.com/fogleman/gg"
	"github.com/lib/pq" // Import for postgres
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkb"
	"github.com/paulmach/orb/geojson"
	"log"
	"math"
	"os"
	"strconv"
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
	// for i := 0; i <= 1; i++ {
	// 	for j := 0; j <= 1; j++ {
	// 		downloadFile(fmt.Sprintf("image_%s_%s.jpeg", fmt.Sprint(i), fmt.Sprint(j)), fmt.Sprintf("%s/1/%s/%s.jpeg", URLPrefix, fmt.Sprint(i), fmt.Sprint(j)))
	// 	}
	// }

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

func TransformBbox(bbox_ []float64) (bbox [4]float64) {
	for i, value := range bbox_ {
		bbox[i] = value
	}
	return
}

// fetch line strings from db by ids
func test_line_wkt() (error, error) {

	var line orb.LineString
	// open connection
	log.Println(psqlConnectionString())
	db, err := sql.Open("postgres", psqlConnectionString())
	if err != nil {
		log.Println("DB Error connection failed")
		log.Println(err)
		return nil, err
	}
	defer db.Close()
	// execute query
	rows, err := db.Query("SELECT ST_AsBinary(line_wkt),bbox from flight where id='4'")

	// MergeImage()
	// MergeImage4_4()

	for rows.Next() {
		// Array for postgres query
		arr := pq.Float64Array{}

		// parse to ST_AsBinary(line_wkt) and bbox to arr
		err := rows.Scan(wkb.Scanner(&line), &arr)
		if err != nil {
			panic(err)
		}

		// Cast postgres array to native go array
		bbox := TransformBbox([]float64(arr))
		log.Println(bbox)
		ImageFlight := NewImage(bbox)
		ImageFlight.FindTiles()
		ImageFlight.TilesAlignment()
		ImageFlight.DownloadTiles()
		ImageFlight.ComposeImage("Flight")

		feature := geojson.NewFeature(line)

		// Convert to lineString (the syntax Geometry. is necessary due to the interface)
		line := feature.Geometry.(orb.LineString)
		// open image
		im, err := gg.LoadPNG("images/Flight_merged.png")
		if err != nil {
			panic(err)
		}
		dc := gg.NewContextForImage(im)
		var longShift = float64(ImageFlight.Images[0][0])
		var latShift = float64(ImageFlight.Images[0][1])
		log.Println(longShift, latShift)
		var ZoomLevel = math.Pow(2, float64(ImageFlight.Tiles[0].Z))
		var TileSize = 512.0
		for _, value := range line {
			var lonBER = value[0] * math.Pi / 180
			var latBER = value[1] * math.Pi / 180
			dc.DrawCircle(LongToPixel(lonBER)*ZoomLevel-TileSize*longShift, LatToPixel(latBER)*ZoomLevel-TileSize*latShift, 4.0)
			dc.Stroke()
			dc.SetRGB(45.0/256.0, 85.0/256.0, 166.0/256.0)
			dc.Fill()
		}

		dc.SavePNG(fmt.Sprintf("images/%s_merged_painted.png", "Flight_Test"))
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return nil, nil
}
