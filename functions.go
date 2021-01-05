package main

import (
	"fmt"
	"github.com/fogleman/gg"
	"io"
	"log"
	"math"
	"net/http"
	"os"
)

// IntMin returns the minimum value
func IntMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type Tile struct {
	Z    int16
	X    int16
	Y    int16
	Lat  float64
	Long float64
}

// Abs returns the absolute value for an unsigned integer
func Abs(x int16) int16 {
	if x < 0 {
		return -x
	}
	return x
}

type Conversion interface {
	deg2num(t *Tile) (x int, y int)
	num2deg(t *Tile) (lat float64, long float64)
}

// Distance returns the added absolute 'distance' between two tiles
// the term distance is not refering to the geographical distance
func (t *Tile) Distance(ref *Tile) (x int16) {
	return Abs(t.X-ref.X) + Abs(t.Y-ref.Y)
}

// Deg2num returns the tiles position x and y
func (t *Tile) Deg2num() (x int16, y int16) {
	x = int16(math.Floor((t.Long + 180.0) / 360.0 * (math.Exp2(float64(t.Z)))))
	y = int16(math.Floor((1.0 - math.Log(math.Tan(t.Lat*math.Pi/180.0)+1.0/math.Cos(t.Lat*math.Pi/180.0))/math.Pi) / 2.0 * (math.Exp2(float64(t.Z)))))
	return
}

// Num2deg returns the latitude and longitude of the upper left corner of the tile
// this function is a method and is called therefore on a tile struct itself
func (t *Tile) Num2deg() (lat float64, long float64) {
	n := math.Pi - 2.0*math.Pi*float64(t.Y)/math.Exp2(float64(t.Z))
	lat = 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
	long = float64(t.X)/math.Exp2(float64(t.Z))*360.0 - 180.0
	return lat, long
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

func MergeImage4_4() {
	const NX int = 2
	const NY int = 2
	var zoom_level int = 2
	// zoom_level = 2
	const ZoomLevelExponent int = 2
	zoom_level = int(math.Pow(2, float64(ZoomLevelExponent)))
	log.Println(zoom_level)
	// k := 1
	for tile_x := 0; tile_x <= ZoomLevelExponent; tile_x++ {
		for tile_y := 0; tile_y <= ZoomLevelExponent; tile_y++ {
			fmt.Println(tile_x, tile_y)
		}
	}

	im, err := gg.LoadJPG("images/0_0.jpg")
	if err != nil {
		panic(err)
	}
	w := im.Bounds().Size().X
	h := im.Bounds().Size().Y
	dc := gg.NewContext(w*2, h*2)
	dc.DrawImage(im, 0*w, 0*h)
	im2, err := gg.LoadJPG("images/1_0.jpg")
	if err != nil {
		panic(err)
	}
	dc.DrawImage(im2, 1*w, 0*h)
	im3, err := gg.LoadJPG("images/0_1.jpg")
	if err != nil {
		panic(err)
	}
	dc.DrawImage(im3, 0*w, 1*h)
	im4, err := gg.LoadJPG("images/1_1.jpg")
	if err != nil {
		panic(err)
	}
	dc.DrawImage(im4, 1*w, 1*h)
	dc.SavePNG("images/merged.png")
}

func downloadFile(filepath string, url string) (err error) {

	// Create the file
	const path string = "images"
	// ignore errors, while creating images folder
	_ = os.Mkdir(path, 0777)
	out, err := os.Create(fmt.Sprintf("%s/%s", path, filepath))
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
