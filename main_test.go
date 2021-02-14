package main

import (
	"fmt"

	// "github.com/fogleman/gg"
	_ "image"
	_ "image/png"
	_ "log"
	"os"
	"testing"
)

type TestCase struct {
	bbox      [4]float64
	ZoomLevel int16
	Name      string
}

// CheckCase simplifies the testing of the different test cases and reduces code duplicity
func (Im *Image) CheckZoomLevel(Z int16, t *testing.T) {
	if Im.Tiles[0].Z != Z && Im.Tiles[1].Z != Z {
		t.Errorf("Zoom Levels are not matching, LeftTile %d, RightTile %d Expectation %d", Im.Tiles[0].Z, Im.Tiles[1].Z, Z)
	}
}

func (Im *Image) CheckNoImages(NoImages int16, t *testing.T) {
	if Im.NoImages != NoImages {
		t.Errorf("NoImages is not matching %d", Im.NoImages)
	}
}

// func CheckKey(Key int16, ExpectedKey int16, t *testing.T) {
// 	if Key != ExpectedKey {
// 		t.Errorf("NoImages is not matching %d", Im.NoImages)
// 	}
// }

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func ReadImage(FileName string) *os.File {
	Image, err := os.Open(FileName)
	CheckError(err)
	return Image
}

func CheckImages(ImageName string) {
	ImageCurrent := ReadImage(fmt.Sprintf("images/%s.png", ImageName))
	ImageReference := ReadImage(fmt.Sprintf("images/%s_Ref.png", ImageName))

	b1 := make([]byte, 64)
	n1, err := ImageCurrent.Read(b1)
	CheckError(err)
	// log.Printf("%d bytes: %s\n", n1, string(b1[:n1]))

	b2 := make([]byte, 64)
	n2, err := ImageReference.Read(b2)
	CheckError(err)

	if string(b1[:n1]) != string(b2[:n2]) {
		panic(fmt.Sprintf("Images are not identical: %s", ImageName))
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
	// Berlin - New York
	CaseBNY := TestCase{[4]float64{-74.006015, 40.71272, 13.38886, 52.517037}, 2, "Berlin - New York"}
	ImageBNY := NewImage(CaseBNY.bbox)

	// Find Tiles including the zoom level
	ImageBNY.FindTiles()
	ImageBNY.CheckZoomLevel(2, t)
	key := ImageBNY.TilesAlignment()
	if key != 0 {
		t.Errorf("Start key of tiles ordering is wrong %d", key)
	}
	ImageBNY.CheckNoImages(2, t)
	ImageBNY.FindBBox()
	ImageBNY.DownloadTiles()
	ImageBNY.ComposeImage("BerlinNewYork")
	CheckImages("BerlinNewYork_merged")
	ImageBNY.DrawImage(&CaseBNY.bbox, "BerlinNewYork")

	// Berlin - Rio Case
	CaseBRIO := TestCase{[4]float64{-43.209373, -22.911014, 13.38886, 52.517037}, 2, "Berlin - RIO"}
	ImageBRIO := NewImage(CaseBRIO.bbox)
	// Find Tiles including the zoom level
	ImageBRIO.FindTiles()
	ImageBRIO.CheckZoomLevel(2, t)
	key = ImageBRIO.TilesAlignment()
	if key != 1 {
		t.Errorf("Start key of tiles ordering is wrong %d", key)
	}
	ImageBRIO.CheckNoImages(4, t)
	ImageBRIO.DownloadTiles()
	ImageBRIO.FindBBox()
	ImageBRIO.ComposeImage("BerlinRio")
	CheckImages("BerlinRio_merged")
	ImageBRIO.DrawImage(&CaseBRIO.bbox, "BerlinRio")

	CaseBHAM := TestCase{[4]float64{10.000654, 52.517037, 13.38886, 53.550341}, 7, "Berlin - Hamburg"}
	ImageBHAM := NewImage(CaseBHAM.bbox)
	// Find Tiles including the zoom level
	ImageBHAM.FindTiles()
	ImageBHAM.CheckZoomLevel(7, t)
	key = ImageBHAM.TilesAlignment()
	if key != 0 {
		t.Errorf("Start key of tiles ordering is wrong %d", key)
	}
	ImageBHAM.CheckNoImages(2, t)
	ImageBHAM.DownloadTiles()
	ImageBHAM.FindBBox()
	ImageBHAM.ComposeImage("BerlinHAM")
	CheckImages("BerlinHAM_merged")
	ImageBHAM.DrawImage(&CaseBHAM.bbox, "BerlinHAM")

	CaseBBARC := TestCase{[4]float64{-8.6107884, 41.1494512, 13.38886, 52.517037}, 4, "Berlin - Barcelona"}
	ImageBBARC := NewImage(CaseBBARC.bbox)
	// Find Tiles including the zoom level
	ImageBBARC.FindTiles()
	ImageBBARC.CheckZoomLevel(4, t)
	key = ImageBBARC.TilesAlignment()
	if key != 0 {
		t.Errorf("Start key of tiles ordering is wrong %d", key)
	}
	ImageBBARC.CheckNoImages(2, t)
	ImageBBARC.DownloadTiles()
	ImageBBARC.FindBBox()
	ImageBBARC.ComposeImage("BerlinBBARC")
	CheckImages("BerlinBBARC_merged")
	ImageBBARC.DrawImage(&CaseBBARC.bbox, "BerlinBBARC")

	// bbox = min Longitude , min Latitude , max Longitude , max Latitude
	CaseFlightFFM := TestCase{[4]float64{8.682127, 50.110922, 8.7667933, 50.8021728}, 7, "Flight around Frankfurt am Main"}
	ImageFlightFFM := NewImage(CaseFlightFFM.bbox)
	// Find Tiles including the zoom level
	ImageFlightFFM.FindTiles()
	ImageFlightFFM.CheckZoomLevel(8, t)
	key = ImageFlightFFM.TilesAlignment()
	if key != 0 {
		t.Errorf("Start key of tiles ordering is wrong %d", key)
	}
	ImageFlightFFM.CheckNoImages(2, t)
	ImageFlightFFM.DownloadTiles()
	ImageFlightFFM.FindBBox()
	ImageFlightFFM.ComposeImage("FlightFFM")
	CheckImages("FlightFFM_merged")
	ImageFlightFFM.DrawImage(&CaseFlightFFM.bbox, "FlightFFM")
}
