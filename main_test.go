package main

import (
	"fmt"
	"github.com/fogleman/gg"
	"log"
	"testing"
)

type TestCase struct {
	bbox      [4]float64
	ZoomLevel int16
	Name      string
}

// FindTiles returns the tiles tht have a distance of one or two to each other
func FindTiles(bbox *[4]float64) (Level *Tile, Level2 *Tile) {
	TileLeft := Tile{11, 0, 0, bbox[1], bbox[0]}
	TileRight := Tile{11, 0, 0, bbox[3], bbox[2]}
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
	log.Println(TileLeft.X, TileLeft.Y, "|", TileRight.X, TileRight.Y)
	return &TileLeft, &TileRight
}

func CreateImage(bbox [4]float64) {
	var WidthHeight = make(map[int16][2]int)
	WidthHeight[0] = [2]int{0, 0}
	WidthHeight[1] = [2]int{0, 1}
	WidthHeight[2] = [2]int{1, 0}
	WidthHeight[3] = [2]int{1, 1}
	TileLeft, TileRight := FindTiles(&bbox)
	Im, RootKey := TileLeft.Download(TileRight)
	log.Println("before", RootKey)
	DownloadTiles(Im, TileLeft.Z)
	log.Println(RootKey)
	im, err := gg.LoadJPG(fmt.Sprintf("images/%d_%d.jpeg", Im.Images[RootKey][0], Im.Images[RootKey][1]))
	if err != nil {
		panic(err)
	}
	w := im.Bounds().Size().X
	h := im.Bounds().Size().Y
	fmt.Println("Creating new image with", int(Im.NoImages))
	dc := gg.NewContext(w*int(Im.NoImages), h*int(Im.NoImages))
	dc.DrawImage(im, WidthHeight[RootKey][1]*w, WidthHeight[RootKey][0]*h)
	// dc.DrawCircle(p.Lon()*512+10, (1-p.Lat())*512, 1.0)
	dc.SavePNG("images/merged_1.png")
	for k, value := range Im.Images {
		if k != RootKey {
			log.Println("Loading", value)
			im, err := gg.LoadJPG(fmt.Sprintf("images/%d_%d.jpeg", value[0], value[1]))
			if err != nil {
				panic(err)
			}
			dc.DrawImage(im, WidthHeight[k][1]*w, WidthHeight[k][0]*h)
		}
	}
	dc.SavePNG("images/merged.png")
}

// CheckCase simplifies the testing of the different test cases and reduces code duplicity
func CheckCase(TestBBox TestCase, t *testing.T) {
	TileLeft, TileRight := FindTiles(&TestBBox.bbox)
	if TileLeft.Z != TileRight.Z {
		t.Errorf("Zoom Levels are not matching, LeftTile %d, RightTile %d", TileLeft.Z, TileRight.Z)
	}
	if TileLeft.Z != TestBBox.ZoomLevel {
		t.Errorf("Test Case %s Expected Zoom Level is wrong: %d", TestBBox.Name, TileLeft.Z)
	}
}

func TestFindTiles(t *testing.T) {

	// BBox consist out of Berlin and New York
	/*
		If we consider the first zoom level we have four tiles.
		Using the coordinates of Berlin and New York, we should get the
		tiles 0 and 1 as the matching tiles
		┌───────┐                         ┌───────┐
		│  New  │ ◀────┐             ┌───▶│Berlin │
		│ York  │      │  ┌────┬────┐│    └───────┘
		└───────┘      └──┼──  │  ──┼┘
		                  │  0 │ 1  │
		                  ├────┼────┤
		                  │    │    │
		                  │  2 │ 3  │
		                  └────┴────┘

	*/
	// Coordinates based on https://www.gps-coordinates.net/
	// Setup of different test cases to find zoom level
	CaseBNY := TestCase{[4]float64{-74.006015, 40.71272, 13.38886, 52.517037}, 2, "Berlin - New York"}
	CheckCase(CaseBNY, t)
	CreateImage(CaseBNY.bbox)

	CaseBRIO := TestCase{[4]float64{-43.209373, -22.911014, 13.38886, 52.517037}, 2, "Berlin - RIO"}
	CheckCase(CaseBRIO, t)
	CreateImage(CaseBRIO.bbox)
	CaseBHAM := TestCase{[4]float64{10.000654, 52.517037, 13.38886, 53.550341}, 7, "Berlin - Hamburg"}
	CheckCase(CaseBHAM, t)
	CreateImage(CaseBHAM.bbox)

	CaseBBARC := TestCase{[4]float64{-8.6107884, 41.1494512, 13.38886, 52.517037}, 4, "Berlin - Barcelona"}
	CheckCase(CaseBBARC, t)
	CaseBR := TestCase{[4]float64{12.482932, 41.89332, 13.38886, 52.517037}, 5, "Berlin - Rome"}
	CheckCase(CaseBR, t)
	// bbox = min Longitude , min Latitude , max Longitude , max Latitude
	CaseFlightFFM := TestCase{[4]float64{8.43103, 50.17878, 10.93463, 50.61335}, 7, "Flight around Frankfurt am Main"}
	CheckCase(CaseFlightFFM, t)

}
