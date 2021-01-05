package main

import (
	"log"
	"testing"
)

func FindZoomLevel(bbox *[4]float64) (Level uint32) {
	TileLeft := new(Tile)
	TileRight := new(Tile)
	TileLeft.Lat = bbox[0]
	TileLeft.Long = bbox[1]
	TileRight.Lat = bbox[2]
	TileRight.Long = bbox[3]
	for z := 0; z < 11; z++ {
		TileLeft.X, TileLeft.Y = TileLeft.Deg2num()
		TileRight.X, TileRight.Y = TileRight.Deg2num()
		distance := TileLeft.Distance(TileRight)
		log.Println(distance)
		if distance == 0 {
			TileLeft.Z++
			TileLeft.Y++
		} else if distance == 1 || distance == 2 {
			break
		}
	}
	return TileLeft.Z
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
	// 						Lat Ber    Long Ber   Lat NY    Long NY
	log.Println("Case Berlin - New York")
	TestBBox := [4]float64{52.517037, 40.712728, 13.38886, -74.006015}
	zoomlevel := FindZoomLevel(&TestBBox)
	// var StartZoomLevel uint32 = 0
	if zoomlevel != 1 {
		t.Errorf("Expected Zoom Level is wrong: %d", zoomlevel)
	}
	// BBox consist out of Berlin and Hamburg
	// replacing coordinates of New York with Berlin
	log.Println("Case Berlin - Hamburg")
	TestBBox[2] = 53.550341
	TestBBox[3] = 10.000654
	zoomlevel = FindZoomLevel(&TestBBox)
	// var StartZoomLevel uint32 = 0
	if zoomlevel != 1 {
		t.Errorf("Expected Zoom Level is wrong: %d", zoomlevel)
	}
}
