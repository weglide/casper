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

func FindTiles(bbox *[4]float64) (Level *Tile, Level2 *Tile) {
	TileLeft := Tile{11, 0, 0, bbox[0], bbox[2]}
	TileRight := Tile{11, 0, 0, bbox[1], bbox[3]}
	for z := 0; z < 11; z++ {
		TileLeft.X, TileLeft.Y = TileLeft.Deg2num()
		TileRight.X, TileRight.Y = TileRight.Deg2num()
		distanceL := TileLeft.Distance(&TileRight)
		log.Println("----------")
		log.Println(TileLeft.X, TileLeft.Y, "|", TileRight.X, TileRight.Y)
		log.Println("DistanceL: ", distanceL, "Zoom", TileLeft.Z)
		if distanceL == 1 || distanceL == 2 {
			break
		} else if distanceL != 0 {
			TileLeft.Z--
			TileRight.Z--
		}
	}
	return &TileLeft, &TileRight
}

func CheckCase(TestBBox TestCase, t *testing.T) {
	TileLeft, TileRight := FindTiles(&TestBBox.bbox)
	if TileLeft.Z != TileRight.Z {
		t.Errorf("Zoom Levels are not matching, LeftTile %d, RightTile %d", TileLeft.Z, TileRight.Z)
	}
	if TileLeft.Z != TestBBox.ZoomLevel {
		t.Errorf("Expected Zoom Level is wrong: %d", TileLeft.Z)
	}
}

func TestNum2deg(t *testing.T) {
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
	// 						Lat Ber  Lat NY    Long Ber    Long NY
	CaseBNY := TestCase{[4]float64{52.517037, 40.712728, 13.38886, -74.006015}, 2, "Berlin - New York"}
	CheckCase(CaseBNY, t)
	CaseBHAM := TestCase{[4]float64{52.517037, 53.550341, 13.38886, 10.000654}, 7, "Berlin - Hamburg"}
	CheckCase(CaseBHAM, t)
	CaseBBARC := TestCase{[4]float64{52.517037, 41.1494512, 13.38886, -8.6107884}, 4, "Berlin - Barcelona"}
	CheckCase(CaseBBARC, t)
}
