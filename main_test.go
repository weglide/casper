package main

import (
	"log"
	"testing"
)

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
	// 						Lat Ber    Long Ber   Lat NY    Long NY
	TestBBox := [4]float64{52.517037, 40.712728, 13.38886, -74.006015}
	log.Println(TestBBox)
	testTile := new(Tile)
	// var StartZoomLevel uint32 = 0
	for z := 0; z < 11; z++ {
		testTile.Lat = TestBBox[0]
		testTile.Long = TestBBox[1]
		var a, b = testTile.Deg2num()
		if a == b {
			testTile.Z++
		} else {
			break
		}
	}
	if testTile.Z != 1 {
		t.Errorf("Expected Zoom Level is wrong")
	}

	log.Println(testTile.Z)
}
