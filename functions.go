package main

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sync"

	"github.com/fogleman/gg"
	"github.com/oliamb/cutter"
)

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
func MaxFloat(x float64, y float64) float64 {
	if x > y {
		return x
	} else {
		return y
	}
}

// IntMin returns the minimum value
func IntMin(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

type Conversion interface {
	deg2num(t *Tile) (x int, y int)
	num2deg(t *Tile) (lat float64, long float64)
}

// Image contains the necessary information to structure to create the image
type Image struct {
	Distance       int16
	StartIndex     int16
	NoImages       int16
	NoImagesWidth  int16
	NoImagesHeight int16
	Images         map[int16][2]int
	bbox           [4]float64
	bboxImage      [4]float64
	RootTile       Tile
}

// FindTiles returns the tiles tht have a distance of one or two to each other
func (Im *Image) FindTiles() {
	// Creating tiles based on the bbox with the finest zoom level
	TileLeft := Tile{11, 0, 0, Im.bbox[1], Im.bbox[0]}
	TileRight := Tile{11, 0, 0, Im.bbox[3], Im.bbox[2]}
	for z := 0; z <= 11; z++ {
		TileLeft.X, TileLeft.Y = TileLeft.Deg2num()
		TileRight.X, TileRight.Y = TileRight.Deg2num()
		distanceX, distanceY := TileLeft.Distance(&TileRight)
		// stop the algorithm if the distance is smaller than 1
		if distanceX == 0 && distanceY == 0 {
			break
			// the zoom level has to be reduced if the distance is still larger than 1
		} else if distanceX >= 1 || distanceY >= 1 {
			TileLeft.Z--
			TileRight.Z--
		}
	}
	Im.RootTile = TileLeft
}

// NewImage is a custom constructor image struct
func NewImage(bbox [4]float64) (Im *Image) {
	Im = new(Image)
	Im.bbox = bbox
	return
}

func (Im *Image) ComposeImage(prefix string) {
	// WidthHeight maps the tiles ordering to the shift of hight and width
	WidthHeight := map[int16][2]int{0: [2]int{0, 0}, 1: [2]int{0, 1}, 2: [2]int{1, 0}, 3: [2]int{1, 1}}
	log.Println(WidthHeight)

	// Load the image for the top left corner
	ImageComposed, err := gg.LoadJPG(fmt.Sprintf("images/%d_%d.jpeg", Im.Images[0][0], Im.Images[0][1]))
	if err != nil {
		panic(err)
	}
	// Width and Height of Image
	w, h := ImageComposed.Bounds().Size().X, ImageComposed.Bounds().Size().Y
	// Standard Case two images
	dc := gg.NewContext(w*int(Im.NoImagesWidth), h*int(Im.NoImagesHeight))

	// Drawing context with 4 images -> 2 Images per Direction
	dc = gg.NewContext(w*2, h*2)
	// Draw Image top left corner
	dc.DrawImage(ImageComposed, WidthHeight[0][1]*w, WidthHeight[0][0]*h)
	for k, value := range Im.Images {
		if k != 0 && value[0] != -1 && value[1] != -1 {
			// log.Println("Loading", value)
			im, err := gg.LoadJPG(fmt.Sprintf("images/%d_%d.jpeg", value[0], value[1]))
			if err != nil {
				panic(err)
			}
			// log.Println("Shift", WidthHeight[k][1]*w, WidthHeight[k][0]*h)
			dc.DrawImage(im, WidthHeight[k][1]*w, WidthHeight[k][0]*h)
		}
	}
	dc.SavePNG(fmt.Sprintf("images/%s_merged.png", prefix))
}

// DownloadTiles saves the required tiles to the folder images
func DownloadTiles(array map[int64][2]int16, Z int16) {
	var wg sync.WaitGroup
	wg.Add(len(array))
	for _, value := range array {
		// Download tiles in parallel
		if value[0] != -1 && value[1] != -1 {
			go func(value [2]int16) {
				downloadFile(fmt.Sprintf("%d_%d", value[0], value[1]), fmt.Sprintf("https://maptiles.glidercheck.com/hypsometric/%d/%d/%d.jpeg", Z, value[0], value[1]))
				defer wg.Done()
			}(value)
		}
	}
	wg.Wait()
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

// Deg2num returns the tiles position x and y
func Deg2num(long float64, lat float64, z int16) (x int16, y int16) {
	x = int16(math.Floor((long + 180.0) / 360.0 * (math.Exp2(float64(z)))))
	y = int16(math.Floor((1.0 - math.Log(math.Tan(lat*math.Pi/180.0)+1.0/math.Cos(lat*math.Pi/180.0))/math.Pi) / 2.0 * (math.Exp2(float64(z)))))
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

// TilesDownload returns the latitude and longitude of the upper left corner of the tile
// this function is a method and is called therefore on a tile struct itself
func TilesDownload(X int16, Y int16, Z int16) (array map[int64][2]int16, ZoomIncrease int16) {

	// Init array of tiles
	array = make(map[int64][2]int16)

	// Check Maximum Level
	ZoomIncrease = 2
	MaxLevel := int16(11)
	if MaxLevel-ZoomIncrease < Z {
		ZoomIncrease = 11 - Z
	}
	index := 0
	/* The assumption is that we have 4 tiles in each direction of the image this leads to
	16 images in total. To determine the X and Y label of each tile we need a nested loop
	in both directions. X and Y are determined similar to Num2deg but with 0.25 steps.
	Afterwards we can use Deg2num to get X and Y.
	*/
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			n := math.Pi - 2.0*math.Pi*float64(float64(Y)+0.25*float64(j))/math.Exp2(float64(Z))
			lat := 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
			long := float64(float64(X)+0.25*float64(i))/math.Exp2(float64(Z))*360.0 - 180.0
			x, y := Deg2num(long, lat, Z+ZoomIncrease)
			array[int64(index)] = [2]int16{x, y}
			index++
		}
	}
	return
}

// Num2deg without creating tile
func Num2deg(X int, Y int, Z int) (lat float64, long float64) {
	n := math.Pi - 2.0*math.Pi*float64(Y)/math.Exp2(float64(Z))
	lat = 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
	long = float64(X)/math.Exp2(float64(Z))*360.0 - 180.0
	return lat, long
}

// DegreeToRadian self explaining
func DegreeToRadian(degree float64) (radian float64) {
	return degree * math.Pi / 180.0
}

// LatLontoXY converts the coordinates (given in degree) to the pixel coordinates
func LatLontoXY(tile_size float64, lat_center float64, lon_center float64, zoom float64) (lon float64, lat float64) {
	C := (tile_size / (2 * math.Pi)) * math.Pow(2, zoom)
	lon = C * (DegreeToRadian(lon_center) + math.Pi)
	lat = C * (math.Pi - math.Log(math.Tan((math.Pi/4)+DegreeToRadian(lat_center)/2)))
	return
}

// DrawImage creates the image for the Test cases in main_Test
func (Im *Image) DrawImage(bbox *[4]float64, array map[int64][2]int16, ZoomIncrease int16, prefix string, RootTileX int16, RootTileY int16) {

	im, err := gg.LoadPNG(fmt.Sprintf("images/%s_merged.png", prefix))
	if err != nil {
		panic(err)
	}
	dc := gg.NewContextForImage(im)

	// var ZoomLevel = math.Pow(2, float64(Im.RootTile.Z))
	var TileSize = 2048.0
	// As the bbox starts with the minimum lat and lon coordinates the variable is namend Min
	LonMinpixel, LatMinpixel := LatLontoXY(TileSize, bbox[1], bbox[0], float64(ZoomIncrease))
	LonMaxpixel, LatMaxpixel := LatLontoXY(TileSize, bbox[3], bbox[2], float64(ZoomIncrease))

	/* The calculated Pixelvalues are equal to the values if the all tiles of this Zoom Level
	are put into one image. Therefore, the top left corner of this image needs to be subtracted.
	*/
	LonMinpixel -= TileSize * float64(RootTileX)
	LatMinpixel -= TileSize * float64(RootTileY)
	LonMaxpixel -= TileSize * float64(RootTileX)
	LatMaxpixel -= TileSize * float64(RootTileY)

	// Draw the circles of the bbox locations
	dc.DrawCircle(LonMinpixel, LatMinpixel, 5.0)
	dc.DrawCircle(LonMaxpixel, LatMaxpixel, 5.0)
	dc.SetLineWidth(2)

	// Set Connection Line
	dc.DrawLine(LonMinpixel, LatMinpixel, LonMaxpixel, LatMaxpixel)
	dc.Stroke()
	dc.SetRGB(0, 0, 0)

	// Save JPEG
	dc.SaveJPG(fmt.Sprintf("images/%s_merged_painted.jpg", prefix), 10)

	// Cropping
	// Calculation of minimum lat and lon, this determines the top left corner based on the bbox
	minLon := math.Min(LonMinpixel, LonMaxpixel) * 0.8
	minLat := math.Min(LatMinpixel, LatMaxpixel) * 0.8
	maxLon := math.Max(LonMinpixel, LonMaxpixel) * 1.1
	maxLat := math.Max(LatMinpixel, LatMaxpixel) * 1.1

	// we need a bbox that is a little bit larger than the current one
	distanceX := math.Abs(maxLon - minLon)
	distanceY := math.Abs(maxLat - minLat)

	log.Println("Distance X", distanceX, "Distance Y", distanceY)
	maxdistance := int(MaxFloat(distanceX, distanceY))

	// if the required distances is smaller than 480 than we want to use at least 48ß
	// TODO: this calculation could be improved because the Anchor Point could be shifted for a better image
	if maxdistance < 480 {
		maxdistance = 480
	}
	croppedImg, err := cutter.Crop(dc.Image(), cutter.Config{
		Width:  maxdistance,
		Height: maxdistance,
		Anchor: image.Point{int(minLon), int(minLat)},
	})
	fo, err := os.Create(fmt.Sprintf("images/%s_merged_painted.png", prefix))
	err = png.Encode(fo, croppedImg)
}

func CreateImage(tiles map[int64][2]int16, prefix string) {
	ImageComposed, err := gg.LoadJPG(fmt.Sprintf("images/%d_%d.jpeg", tiles[0][0], tiles[0][1]))
	if err != nil {
		panic(err)
	}

	// Width and Height of Image
	w, h := ImageComposed.Bounds().Size().X, ImageComposed.Bounds().Size().Y
	// Standard Case two images
	dc := gg.NewContext(w*int(4), h*int(4))

	// Drawing context with 4 images -> 2 Images per Direction
	// Draw Image top left corner
	dc.DrawImage(ImageComposed, 0, 0)
	CounterWidth := 0
	CounterHeight := 0
	// for k, value := range tiles {
	for k := 0; k < 16; k++ {
		// log.Println("Loading", k)
		im, err := gg.LoadJPG(fmt.Sprintf("images/%d_%d.jpeg", tiles[int64(k)][0], tiles[int64(k)][1]))
		log.Println(tiles[int64(k)][0], tiles[int64(k)][1])
		if err != nil {
			panic(err)
		}
		// log.Println("Shift", CounterWidth*w, CounterHeight*h)
		dc.DrawImage(im, CounterWidth*w, CounterHeight*h)
		CounterHeight++
		if (k+1)%4 == 0 && k >= 1 {
			// log.Println("shifting")
			CounterWidth++
			CounterHeight = 0
		}
	}
	dc.SavePNG(fmt.Sprintf("images/%s_merged.png", prefix))
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
	out, err := os.Create(fmt.Sprintf("%s/%s.jpeg", path, filepath))
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
