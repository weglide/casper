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
	TestBBox := [4]float32{52.517037, 40.712728, 13.38886, -74.006015}
	log.Println(TestBBox)
	testTile := new(Tile)
	testTile.X = 10
	testTile.Z = 10
	log.Println(testTile.Num2deg())
}
