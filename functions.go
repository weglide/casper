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
	Tiles          [2]Tile
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
		// log.Println(distanceX, distanceY, z, TileRight.X, TileRight.Y, TileLeft.X, TileLeft.Y)
		// stop the algorithm if the distance is smaller than 1
		if distanceX == 0 && distanceY == 0 {
			log.Println(distanceX, distanceY, z, TileRight.X, TileRight.Y, TileLeft.X, TileLeft.Y)
			log.Println(TileLeft)
			break
			// the zoom level has to be reduced if the distance is still larger than 1
		} else if distanceX >= 1 || distanceY >= 1 {
			TileLeft.Z--
			TileRight.Z--
		}
	}
	Im.Tiles[0] = TileLeft
	Im.Tiles[1] = TileRight
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
			log.Println("Loading", value)
			im, err := gg.LoadJPG(fmt.Sprintf("images/%d_%d.jpeg", value[0], value[1]))
			if err != nil {
				panic(err)
			}
			log.Println("Shift", WidthHeight[k][1]*w, WidthHeight[k][0]*h)
			dc.DrawImage(im, WidthHeight[k][1]*w, WidthHeight[k][0]*h)
		}
	}
	dc.SavePNG(fmt.Sprintf("images/%s_merged.png", prefix))
}

func (Im *Image) FindBBox() {
	/* Shifting  is required in order to get the right lat and lon coordinates from the tiles
	In the default case the values of upper left corner are returned. This doesn't work for the tile ordering.
	*/
	Shifting := map[int16][2]int{1: [2]int{1, 0}, 2: [2]int{0, 1}, 3: [2]int{1, 1}}
	// min Longitude , min Latitude , max Longitude , max Latitude
	var LatMin, LatMax, LonMin, LonMax float64
	for k, value := range Im.Images {
		if k == 0 && value[0] != -1 && value[1] != -1 {
			LatMin, LonMin = Num2deg(value[0], value[1], int(Im.Tiles[0].Z))
			LatMax = LatMin
			LonMax = LonMin
		} else if value[0] != -1 && value[1] != -1 {
			lat, lon := Num2deg(value[0]+Shifting[k][0], value[1]+Shifting[k][1], int(Im.Tiles[0].Z))
			log.Printf("Lat %f Lon %f", lat, lon)
			LatMin = math.Min(LatMin, lat)
			LatMax = math.Max(LatMax, lat)
			LonMin = math.Min(LonMin, lon)
			LonMax = math.Max(LonMax, lon)
		}
	}
	Im.bboxImage[0] = LonMin
	Im.bboxImage[1] = LatMin
	Im.bboxImage[2] = LonMax
	Im.bboxImage[3] = LatMax
}

// TilesAlignment determines positioning of the tiles to be downloaded
func (Im *Image) TilesAlignment() (RootKey int16) {
	t := Im.Tiles[0]
	ref := Im.Tiles[1]
	distanceX, distanceY := t.Distance(&ref)
	Im.Distance = distanceX + distanceY
	// tLat, tLon := t.Num2deg()
	// refLat, refLon := ref.Num2deg()
	// Default case
	Im.StartIndex = 0
	Im.NoImages = 2
	// per default all images have the value -1, -1, later we can check by this standard if the images need to be downloaded or not
	// the default case with 0,0 for each entry is not suitable, as we could have those tiles
	Im.Images = map[int16][2]int{0: {int(Im.Tiles[0].X), int(Im.Tiles[0].Y)}, 1: {int(Im.Tiles[0].X + 1), int(Im.Tiles[0].Y)}, 2: {int(Im.Tiles[0].X), int(Im.Tiles[0].Y + 1)}, 3: {int(Im.Tiles[0].X + 1), int(Im.Tiles[0].Y + 1)}}
	Im.NoImages = 4
	RootKey = 0
	// if Im.Distance == 1 {
	// 	// two tiles differ horizontally but are vertically identical
	// 	if tLon == refLon {
	// 		// Case 1
	// 		Im.NoImagesHeight = 2
	// 		Im.NoImagesWidth = 1
	// 		if tLat > refLat {
	// 			/* Tiles Ordering
	// 			┌─────────┬─────────┐
	// 			│         │         │
	// 			│    0    │    1    │
	// 			│         │         │
	// 			├─────────┼─────────┤
	// 			│         │         │
	// 			│    2    │    3    │
	// 			│         │         │
	// 			└─────────┴─────────┘
	// 			*/
	// 			Im.Images[0] = [2]int{int(t.X), int(t.Y)}
	// 			Im.Images[2] = [2]int{int(ref.X), int(ref.Y)}
	// 			RootKey = 0
	// 			// Case 2
	// 		} else {
	// 			Im.Images[0] = [2]int{int(ref.X), int(ref.Y)}
	// 			Im.Images[2] = [2]int{int(t.X), int(t.Y)}
	// 			RootKey = 0
	// 		}
	// 		// two tiles differ vertically but are horizontally identical
	// 	} else if tLat == refLat {
	// 		Im.NoImagesHeight = 1
	// 		Im.NoImagesWidth = 2
	// 		// Case 3
	// 		if tLon < refLon {
	// 			RootKey = 0
	// 			Im.Images[0] = [2]int{int(t.X), int(t.Y)}
	// 			Im.Images[1] = [2]int{int(ref.X), int(ref.Y)}
	// 			// Case 4
	// 		} else {
	// 			RootKey = 0
	// 			Im.Images[1] = [2]int{int(t.X), int(t.Y)}
	// 			Im.Images[0] = [2]int{int(ref.X), int(ref.Y)}
	// 		}
	// 	}
	// } else if Im.Distance == 2 {
	// 	// four images have to be downloaded
	// 	Im.NoImages = 4
	// 	// two tiles differ horizontally but are vertically identical
	// 	if tLat < refLat {
	// 		// Case 1
	// 		if tLon < refLon {
	// 			Im.StartIndex = 1
	// 			RootKey = 1
	// 			// Images from the calculation
	// 			Im.Images[1] = [2]int{int(ref.X), int(ref.Y)}
	// 			Im.Images[2] = [2]int{int(t.X), int(t.Y)}
	// 			// Additional Images
	// 			Im.Images[0] = [2]int{int(ref.X) - 1, int(ref.Y)}
	// 			Im.Images[3] = [2]int{int(t.X) + 1, int(t.Y)}
	// 			// Case 2
	// 		} else {
	// 			RootKey = 0
	// 			Im.Images[0] = [2]int{int(t.X), int(t.Y)}
	// 			Im.Images[3] = [2]int{int(ref.X), int(ref.Y)}
	// 			// Additional Images
	// 			Im.Images[1] = [2]int{int(t.X) + 1, int(t.Y)}
	// 			Im.Images[2] = [2]int{int(ref.X) - 1, int(ref.Y)}
	// 		}
	// 		// two tiles differ vertically but are horizontally identical
	// 	} else {
	// 		// Case 3
	// 		if tLon < refLon {
	// 			RootKey = 0
	// 			Im.Images[0] = [2]int{int(ref.X), int(ref.Y)}
	// 			Im.Images[3] = [2]int{int(t.X), int(t.Y)}
	// 			// Additional Images
	// 			Im.Images[1] = [2]int{int(ref.X) + 1, int(ref.Y)}
	// 			Im.Images[2] = [2]int{int(t.X) - 1, int(t.Y)}
	// 			// Case 4
	// 		} else if tLon > refLon {
	// 			Im.StartIndex = 1
	// 			RootKey = 1
	// 			Im.Images[1] = [2]int{int(t.X), int(t.Y)}
	// 			Im.Images[2] = [2]int{int(ref.X), int(ref.Y)}
	// 			// Additional Images
	// 			Im.Images[0] = [2]int{int(t.X) - 1, int(t.Y)}
	// 			Im.Images[3] = [2]int{int(ref.X) + 1, int(ref.Y)}
	// 		}
	// 	}
	// }
	return
}

// DownloadTiles saves the required tiles to the folder images
func (Im *Image) DownloadTiles() {
	for _, value := range Im.Images {
		log.Println("Tile value", value)
		if value[0] != -1 && value[1] != -1 {
			log.Printf("Downloading Tiles %d %d with Zoom Level %d", value[0], value[1], Im.Tiles[0].Z+1)
			downloadFile(fmt.Sprintf("%d_%d", value[0], value[1]), fmt.Sprintf("https://maptiles.glidercheck.com/hypsometric/%d/%d/%d.jpeg", Im.Tiles[0].Z+1, value[0], value[1]))
		}
	}
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
func (t *Tile) TilesDownload() (array map[int64][2]int16) {
	// var array [16]float64
	array = make(map[int64][2]int16)
	n := math.Pi - 2.0*math.Pi*float64(t.Y)/math.Exp2(float64(t.Z))
	lat := 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
	long := float64(t.X)/math.Exp2(float64(t.Z))*360.0 - 180.0
	x, y := Deg2num(long, lat, 1)
	array[0] = [2]int16{x, y}
	return
}

// Num2deg without creating tile
func Num2deg(X int, Y int, Z int) (lat float64, long float64) {
	n := math.Pi - 2.0*math.Pi*float64(Y)/math.Exp2(float64(Z))
	lat = 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
	long = float64(X)/math.Exp2(float64(Z))*360.0 - 180.0
	return lat, long
}

func LongToPixel(lon float64) (pix float64) {
	return 512.0 / (2 * math.Pi) * (lon + math.Pi)
}
func LatToPixel(lat float64) (pix float64) {
	return 512.0 / (2 * math.Pi) * (math.Pi - math.Log(math.Tan(math.Pi/4+lat/2)))
}

func (Im *Image) DrawImage(bbox *[4]float64, prefix string) {
	// Rio
	var lonRIO = bbox[0] * math.Pi / 180
	var latRIO = bbox[1] * math.Pi / 180
	// Ber
	var lonBER = bbox[2] * math.Pi / 180
	var latBER = bbox[3] * math.Pi / 180

	// p[1] == p.Lat()
	// Lat
	bbox[1] = (bbox[1] - Im.bboxImage[1]) / (Im.bboxImage[3] - Im.bboxImage[1])
	bbox[3] = (bbox[3] - Im.bboxImage[1]) / (Im.bboxImage[3] - Im.bboxImage[1])
	// Lon
	bbox[0] = (bbox[0] - Im.bboxImage[0]) / (Im.bboxImage[2] - Im.bboxImage[0])
	bbox[2] = (bbox[2] - Im.bboxImage[0]) / (Im.bboxImage[2] - Im.bboxImage[0])

	im, err := gg.LoadPNG(fmt.Sprintf("images/%s_merged.png", prefix))
	if err != nil {
		panic(err)
	}
	dc := gg.NewContextForImage(im)
	var longShift = float64(Im.Images[0][0])
	var latShift = float64(Im.Images[0][1])
	log.Printf("Lon BER %f Lat BER %f Pixel Lon BER %f Pixel Lat BER %f", lonBER, latBER, LongToPixel(lonBER), LatToPixel(latBER))
	log.Printf("Lon RIO %f Lat RIO %f Pixel Lon RIO %f Pixel Lat RIO %f", lonRIO, latRIO, LongToPixel(lonRIO), LatToPixel(lonRIO))
	var ZoomLevel = math.Pow(2, float64(Im.Tiles[0].Z+1))
	var TileSize = 1024.0
	lonBERpixel := LongToPixel(lonBER)*ZoomLevel - TileSize*longShift
	latBERpixel := LatToPixel(latBER)*ZoomLevel - TileSize*latShift
	lonRIOpixel := LongToPixel(lonRIO)*ZoomLevel - TileSize*longShift
	latRIOpixel := LatToPixel(latRIO)*ZoomLevel - TileSize*latShift
	dc.DrawCircle(lonBERpixel, latBERpixel, 5.0)
	dc.DrawCircle(lonRIOpixel, latRIOpixel, 5.0)
	log.Println("lon Ber", lonBERpixel, "lat Ber", latBERpixel, "lon RIO", lonRIOpixel, "lat RIO", latRIOpixel)
	distanceX := math.Abs(lonBERpixel - lonRIOpixel)
	distanceY := math.Abs(latBERpixel - latRIOpixel)
	log.Println("Distance X", distanceX, "Distance Y", distanceY)
	minLon := math.Min(lonBERpixel, lonRIOpixel)
	minLat := math.Min(latBERpixel, latRIOpixel)
	// maxdistance := int(MaxFloat(distanceX, distanceY) * 1.5)
	log.Println(distanceX, distanceY)
	dc.DrawLine(lonBERpixel, latBERpixel, lonRIOpixel, latRIOpixel)
	dc.Stroke()
	dc.SetRGB(0, 0, 0)
	dc.SavePNG(fmt.Sprintf("images/%s_merged_painted.png", prefix))
	AnchorPointLon := int(minLon * 0.5)
	AnchorPointLat := int(minLat * 0.5)
	croppedImg, err := cutter.Crop(dc.Image(), cutter.Config{
		Width:  480,
		Height: 480,
		Anchor: image.Point{AnchorPointLon, AnchorPointLat},
	})
	fo, err := os.Create(fmt.Sprintf("images/%s_merged_painted.png", prefix))
	err = png.Encode(fo, croppedImg)
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
