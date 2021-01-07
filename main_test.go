package main

import (
	"log"
	"testing"
)

func FindTiles(bbox *[4]float64) (Level *Tile, Level2 *Tile) {
	TileLeft := new(Tile)
	TileRight := new(Tile)
	TileLeft.Lat = bbox[0]
	TileLeft.Long = bbox[2]
	TileRight.Lat = bbox[1]
	TileRight.Long = bbox[3]
	TileRight.Z = 11
	TileLeft.Z = 11

	for z := 0; z < 11; z++ {
		TileLeft.X, TileLeft.Y = TileLeft.Deg2num()
		TileRight.X, TileRight.Y = TileRight.Deg2num()
		distanceL := TileLeft.Distance(TileRight)
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
	return TileLeft, TileRight
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
	log.Println("Case Berlin - New York")
	TestBBox := [4]float64{52.517037, 40.712728, 13.38886, -74.006015}
	TileLeft, TileRight := FindTiles(&TestBBox)
	if TileLeft.Z != TileRight.Z {
		t.Errorf("Zoom Levels are not matching, LeftTile %d, RightTile %d", TileLeft.Z, TileRight.Z)
	}
	if TileRight.Z != 2 {
		t.Errorf("Expected Zoom Level is wrong: %d", TileLeft.Z)
	}
	// BBox consist out of Berlin and Hamburg
	// replacing coordinates of New York with Berlin
	log.Println("Case Berlin - Hamburg")
	TestBBox[1] = 53.550341
	TestBBox[3] = 10.000654
	TileLeft, TileRight = FindTiles(&TestBBox)
	if TileLeft.Z != TileRight.Z {
		t.Errorf("Zoom Levels are not matching, LeftTile %d, RightTile %d", TileLeft.Z, TileRight.Z)
	}
	if TileLeft.Z != 7 {
		t.Errorf("Expected Zoom Level is wrong: %d", TileLeft.Z)
	}
	log.Println("Case Berlin - Barcelona")
	TestBBox[1] = 41.1494512
	TestBBox[3] = -8.6107884
	TileLeft, TileRight = FindTiles(&TestBBox)
	if TileLeft.Z != TileRight.Z {
		t.Errorf("Zoom Levels are not matching, LeftTile %d, RightTile %d", TileLeft.Z, TileRight.Z)
	}
	// expected zoom level should be 4
	if TileLeft.Z != 4 {
		t.Errorf("Expected Zoom Level is wrong: %d", TileLeft.Z)
	}

}
