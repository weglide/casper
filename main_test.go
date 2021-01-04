package main

import (
	// "log"
	"testing"
)

func FindZoomLevel(bbox *[4]float64) (Level uint32) {
	testTile := new(Tile)
	for z := 0; z < 11; z++ {
		testTile.Lat = bbox[0]
		testTile.Long = bbox[1]
		var a, b = testTile.Deg2num()
		if a == b {
			testTile.Z++
		} else {
			break
		}
	}
	return testTile.Z
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
	// 						Lat Ber    Long Ber   Lat NY    Long NY
	TestBBox := [4]float64{52.517037, 40.712728, 13.38886, -74.006015}
	zoomlevel := FindZoomLevel(&TestBBox)
	// var StartZoomLevel uint32 = 0
	if zoomlevel != 1 {
		t.Errorf("Expected Zoom Level is wrong: %d", zoomlevel)
	}
}
