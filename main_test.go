package main

import (
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
		// log.Println("----------")
		if distanceX <= 1 && distanceY <= 1 {
			break
		} else if distanceX > 1 || distanceY > 1 {
			TileLeft.Z--
			TileRight.Z--
		}
	}
	log.Println(TileLeft.X, TileLeft.Y, "|", TileRight.X, TileRight.Y)
	// log.Println("DistanceL: ", distanceL, "Zoom", TileLeft.Z)
	return &TileLeft, &TileRight
}

// CheckCase simplifies the testing of the different test cases and reduces code duplicity
func CheckCase(TestBBox TestCase, t *testing.T) {
	TileLeft, TileRight := FindTiles(&TestBBox.bbox)
	Im := TileLeft.Download(TileRight)
	log.Println(Im)
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
	CaseBRIO := TestCase{[4]float64{-43.209373, -22.911014, 13.38886, 52.517037}, 2, "Berlin - RIO"}
	CheckCase(CaseBRIO, t)
	CaseBHAM := TestCase{[4]float64{10.000654, 52.517037, 13.38886, 53.550341}, 7, "Berlin - Hamburg"}
	CheckCase(CaseBHAM, t)
	CaseBBARC := TestCase{[4]float64{-8.6107884, 41.1494512, 13.38886, 52.517037}, 4, "Berlin - Barcelona"}
	CheckCase(CaseBBARC, t)
	CaseBR := TestCase{[4]float64{12.482932, 41.89332, 13.38886, 52.517037}, 5, "Berlin - Rome"}
	CheckCase(CaseBR, t)
	// bbox = min Longitude , min Latitude , max Longitude , max Latitude
	CaseFlightFFM := TestCase{[4]float64{8.43103, 50.17878, 10.93463, 50.61335}, 7, "Flight around Frankfurt am Main"}
	CheckCase(CaseFlightFFM, t)
}
