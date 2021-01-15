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

// Max returns maximum of two values
func Max(x int16, y int16) int16 {
	if x > y {
		return x
	} else {
		return y
	}
}

type Conversion interface {
	deg2num(t *Tile) (x int, y int)
	num2deg(t *Tile) (lat float64, long float64)
}

// Image contains the necessary information to structure to create the image
type Image struct {
	Distance   int16
	StartIndex int16
	NoImages   int16
	Images     map[int16][2]int16
	bbox       [4]float64
	Tiles      [2]Tile
}

// FindTiles returns the tiles tht have a distance of one or two to each other
func (Im *Image) FindTiles() {
	TileLeft := Tile{11, 0, 0, Im.bbox[1], Im.bbox[0]}
	TileRight := Tile{11, 0, 0, Im.bbox[3], Im.bbox[2]}
	for z := 0; z < 11; z++ {
		TileLeft.X, TileLeft.Y = TileLeft.Deg2num()
		TileRight.X, TileRight.Y = TileRight.Deg2num()
		distanceX, distanceY := TileLeft.Distance(&TileRight)
		if distanceX <= 1 && distanceY <= 1 {
			break
		} else if distanceX > 1 || distanceY > 1 {
			TileLeft.Z--
			TileRight.Z--
		}
	}
	Im.Tiles[0] = TileLeft
	Im.Tiles[1] = TileRight

	// log.Println(TileLeft.X, TileLeft.Y, "|", TileRight.X, TileRight.Y)
	// return &TileLeft, &TileRight
}

// NewImage is a custom constructor image struct
func NewImage(bbox [4]float64) (Im *Image) {
	Im = new(Image)
	Im.bbox = bbox
	return
}

func (Im *Image) CreateImage() {
	// WidthHeight maps the tiles ordering to the shift of hight and width
	WidthHeight := map[int16][2]int{0: [2]int{0, 0}, 1: [2]int{0, 1}, 2: [2]int{1, 0}, 3: [2]int{1, 1}}
	log.Println(WidthHeight)
	// TileLeft, TileRight := FindTiles(&Im.bbox)
	// Im, RootKey := TileLeft.Download(TileRight)
	// log.Println("before", RootKey)
	// DownloadTiles(Im, TileLeft.Z)
	// log.Println(RootKey)
	// im, err := gg.LoadJPG(fmt.Sprintf("images/%d_%d.jpeg", Im.Images[RootKey][0], Im.Images[RootKey][1]))
	// if err != nil {
	// 	panic(err)
	// }
	// w := im.Bounds().Size().X
	// h := im.Bounds().Size().Y
	// fmt.Println("Creating new image with", int(Im.NoImages))
	// dc := gg.NewContext(w*int(Im.NoImages), h*int(Im.NoImages))
	// dc.DrawImage(im, WidthHeight[RootKey][1]*w, WidthHeight[RootKey][0]*h)
	// // dc.DrawCircle(p.Lon()*512+10, (1-p.Lat())*512, 1.0)
	// dc.SavePNG("images/merged_1.png")
	// for k, value := range Im.Images {
	// 	if k != RootKey {
	// 		log.Println("Loading", value)
	// 		im, err := gg.LoadJPG(fmt.Sprintf("images/%d_%d.jpeg", value[0], value[1]))
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 		dc.DrawImage(im, WidthHeight[k][1]*w, WidthHeight[k][0]*h)
	// 	}
	// }
	// dc.SavePNG("images/merged.png")

}

// Download determines the tiles to be downloaded
func (Im *Image) Download() (RootKey int16) {
	t := Im.Tiles[0]
	ref := Im.Tiles[1]
	distanceX, distanceY := t.Distance(&ref)
	Im.Distance = distanceX + distanceY
	tLat, tLon := t.Num2deg()
	refLat, refLon := ref.Num2deg()
	// Default case
	Im.StartIndex = 0
	Im.NoImages = 2
	Im.Images = make(map[int16][2]int16)
	if Im.Distance == 1 {
		// two tiles differ horizontally but are vertically identical
		if tLon == refLon {
			// Case 1
			if tLat < refLat {
				/* Tiles Ordering
				┌─────────┬─────────┐
				│         │         │
				│    0    │    1    │
				│         │         │
				├─────────┼─────────┤
				│         │         │
				│    2    │    3    │
				│         │         │
				└─────────┴─────────┘
				*/
				Im.Images[0] = [2]int16{t.X, t.Y}
				Im.Images[2] = [2]int16{ref.X, ref.Y}
				RootKey = 0
				// Case 2
			} else {
				Im.Images[0] = [2]int16{ref.X, ref.Y}
				Im.Images[2] = [2]int16{t.X, t.Y}
				RootKey = 0
			}
			// two tiles differ vertically but are horizontally identical
		} else if tLat == refLat {
			// Case 3
			if tLon < refLon {
				RootKey = 0
				Im.Images[0] = [2]int16{t.X, t.Y}
				Im.Images[1] = [2]int16{ref.X, ref.Y}
				// Case 4
			} else {
				RootKey = 0
				Im.Images[1] = [2]int16{t.X, t.Y}
				Im.Images[0] = [2]int16{ref.X, ref.Y}
			}
		}
	} else if Im.Distance == 2 {
		// four images have to be downloaded
		Im.NoImages = 4
		// two tiles differ horizontally but are vertically identical
		if tLat < refLat {
			// Case 1
			if tLon < refLon {
				Im.StartIndex = 1
				RootKey = 1
				Im.Images[1] = [2]int16{ref.X, ref.Y}
				Im.Images[2] = [2]int16{t.X, t.Y}
				// Case 2
			} else {
				RootKey = 0
				Im.Images[0] = [2]int16{t.X, t.Y}
				Im.Images[3] = [2]int16{ref.X, ref.Y}
			}
			// two tiles differ vertically but are horizontally identical
		} else {
			// Case 3
			if tLon < refLon {
				RootKey = 0
				Im.Images[0] = [2]int16{ref.X, ref.Y}
				Im.Images[3] = [2]int16{t.X, t.Y}
				// Case 4
			} else if tLon > refLon {
				Im.StartIndex = 1
				RootKey = 1
				Im.Images[1] = [2]int16{t.X, t.Y}
				Im.Images[2] = [2]int16{ref.X, ref.Y}
			}
		}
	}
	return
}

// Distance returns the added absolute 'distance' between two tiles
// the term distance is not refering to the geographical distance
func (t *Tile) Distance(ref *Tile) (Distx int16, Disty int16) {
	return Abs(t.X - ref.X), Abs(t.Y - ref.Y)
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

func DownloadTiles(Im *Image, Z int16) {
	for _, value := range Im.Images {
		log.Println("Downloading", value)
		downloadFile(fmt.Sprintf("%d_%d", value[0], value[1]), fmt.Sprintf("https://maptiles.glidercheck.com/hypsometric/%d/%d/%d.jpeg", Z, value[0], value[1]))
	}
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
	log.Println("start creating folder")
	_ = os.Mkdir(path, 0777)
	out, err := os.Create(fmt.Sprintf("%s/%s.jpeg", path, filepath))
	if err != nil {
		panic(err)
	}
	defer out.Close()
	log.Println("end creating folder")

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
