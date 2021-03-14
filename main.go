package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"strconv"

	"github.com/fogleman/gg"
	"github.com/lib/pq"
	"github.com/oliamb/cutter"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkb"
	"github.com/paulmach/orb/geojson"
	"github.com/urfave/cli/v2"

	"os"
)

const (
	ColorScale float64 = 256.0
	ColorRed   float64 = 45.0 / ColorScale
	ColorGreen float64 = 85.0 / ColorScale
	ColorBlue  float64 = 166.0 / ColorScale
	TileSize   float64 = 2048.0
	// e.g. 0.1 means 10 % larger bbox
	BufferforCropping float64 = 0.1
	ImageSize         int     = 480
	URLPrefix         string  = "https://maptiles.glidercheck.com/hypsometric"
)

func main() {
	var (
		FlightID        uint
		CircleThickness float64
		Prefix          string
	)

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:        "id",
				Value:       1,
				Usage:       "Flight ID to pe processed",
				Destination: &FlightID,
			},
			&cli.Float64Flag{
				Name:        "thickness",
				Value:       1.0,
				Aliases:     []string{"th"},
				Usage:       "Thinkness of the line string",
				Destination: &CircleThickness,
			},
			&cli.StringFlag{
				Name:        "prefix",
				Value:       "",
				Aliases:     []string{"p"},
				Usage:       "Prefix for filename",
				Destination: &Prefix,
			},
		},
		Action: func(c *cli.Context) error {
			LOCAL, _ := strconv.ParseBool(os.Getenv("LOCAL"))
			log.Printf("Processing Flight ID %d\n", FlightID)
			// switch between lambda and local environment
			if LOCAL == true {
				PlotFlight(FlightID, CircleThickness, Prefix)
			}
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// fetch line strings from db by ids
func PlotFlight(FlightID uint, CircleThickness float64, Prefix string) error {
	var line orb.LineString
	row := GetRow(FlightID)

	// Array for postgres query
	arr := pq.Float64Array{}
	// parse to ST_AsBinary(line_wkt) and bbox to arr
	err := row.Scan(wkb.Scanner(&line), &arr)
	if err != nil {
		panic(err)
	}

	// Cast postgres array to native go array
	bbox := TransformBbox([]float64(arr))
	ImageFlight := NewImage(bbox)
	// Find Tiles including the zoom level
	ImageFlight.FindRootTile()
	// Determine tiles and download them
	tiles, ZoomIncrease := TilesDownload(ImageFlight.RootTile.X, ImageFlight.RootTile.Y, ImageFlight.RootTile.Z)
	DownloadTiles(tiles, ImageFlight.RootTile.Z+ZoomIncrease)
	CreateImage(tiles, "Flight")

	// Handle GeoJSON Linestring
	feature := geojson.NewFeature(line)
	// Convert to lineString (the syntax Geometry. is necessary due to the interface)
	line = feature.Geometry.(orb.LineString)

	// open image
	im, err := gg.LoadJPG("images/Flight_merged.jpeg")
	if err != nil {
		panic(err)
	}
	dc := gg.NewContextForImage(im)
	longShift := float64(ImageFlight.RootTile.X)
	latShift := float64(ImageFlight.RootTile.Y)

	// Plot each point of the linestring onto the image
	log.Println("Plotting flight")
	for _, value := range line {
		// Obtain the lat and lon values converted into pixels
		lonPixel, latPixel := LatLontoXY(TileSize, value[1], value[0], float64(ImageFlight.RootTile.Z))

		// -TileSize*longShift is necessary in order to shift the origin of the pixels based on the images
		// otherwise lonPixel and latPixel don't match with the canvas
		dc.DrawCircle(lonPixel-TileSize*longShift, latPixel-TileSize*latShift, CircleThickness)
		dc.Stroke()
		dc.SetRGB(ColorRed, ColorGreen, ColorBlue)
		dc.Fill()
	}

	// ----------------- In this section the image will be cropped -----------------
	log.Println("Cropping")
	// Calculate BBOX in pixels
	lonPixelFirst, latPixelFirst := LatLontoXY(TileSize, bbox[1], bbox[0], float64(ImageFlight.RootTile.Z))
	lonPixelSecond, latPixelSecond := LatLontoXY(TileSize, bbox[3], bbox[2], float64(ImageFlight.RootTile.Z))

	// Subtract shifting of tiles
	lonPixelFirst -= TileSize * longShift
	lonPixelSecond -= TileSize * longShift
	latPixelFirst -= TileSize * latShift
	latPixelSecond -= TileSize * latShift

	// Determine the the min nad max vlaues with buffer included
	minLon := math.Min(lonPixelFirst, lonPixelSecond) * (1 - BufferforCropping)
	minLat := math.Min(latPixelFirst, latPixelSecond) * (1 - BufferforCropping)
	maxLon := math.Max(lonPixelFirst, lonPixelSecond) * (1 + BufferforCropping)
	maxLat := math.Max(latPixelFirst, latPixelSecond) * (1 + BufferforCropping)

	// determin "Pixeldistance" in Lat and Lon direction
	distanceX := math.Abs(maxLon - minLon)
	distanceY := math.Abs(maxLat - minLat)
	// Use greater distance for distance
	maxdistance := int(MaxFloat(distanceX, distanceY))
	// the minimum size of an Image
	if maxdistance < ImageSize {
		maxdistance = ImageSize
	}
	croppedImg, err := cutter.Crop(dc.Image(), cutter.Config{
		Width:  maxdistance,
		Height: maxdistance,
		Anchor: image.Point{int(minLon), int(minLat)},
	})
	log.Println("Saving Image")
	fo, err := os.Create(fmt.Sprintf("%sFlight_merged_painted.jpeg", Prefix))
	err = png.Encode(fo, croppedImg)
	return nil
}
