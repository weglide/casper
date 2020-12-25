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
type MinMax struct {
	latMin float64 // = 90.0
	latMax float64 // = -90.0
	lonMin float64 // = 180.0
	lonMax float64 // = -180.0
}

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

func MergeImage() {
	const NX = 4
	const NY = 3
	im, err := gg.LoadPNG("images/out.png")
	if err != nil {
		panic(err)
	}
	w := im.Bounds().Size().X
	h := im.Bounds().Size().Y
	dc := gg.NewContext(w, h*2)
	dc.DrawImage(im, 0*w, 0*h)
	dc.DrawImage(im, 0*w, 1*h)
	dc.SavePNG("overlay.png")
	im2, err := gg.LoadPNG("overlay.png")
	log.Println(im2.Bounds())
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

	for rows.Next() {
		col1arr := []float64{}
		arr := pq.Float64Array{}
		// ...
		err := rows.Scan(wkb.Scanner(&line), &arr)
		col1arr = []float64(arr)
		fmt.Println(col1arr[0])

		if err != nil {
			panic(err)
			return nil, err
		}
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
		dc.SetRGBA(0, 0, 1, 1)
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
