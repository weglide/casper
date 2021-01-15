package main

import (
	// "fmt"
	// "github.com/fogleman/gg"
	// "log"
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

func (Im *Image) CheckNoTiles(NoImages int16, t *testing.T) {
	if Im.NoImages != NoImages {
		t.Errorf("NoImages is not matching %d", Im.NoImages)
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
	ImageBNY := NewImage(CaseBNY.bbox)
	// Find Tiles including the zoom level
	ImageBNY.FindTiles()
	ImageBNY.CheckZoomLevel(2, t)
	CheckNoImages(ImageBNY, 2, t)
	ImageBNY.CreateImage()

	// CheckCase(CaseBNY, t)
	// CreateImage(CaseBNY.bbox)

	// CaseBRIO := TestCase{[4]float64{-43.209373, -22.911014, 13.38886, 52.517037}, 2, "Berlin - RIO"}
	// CheckCase(CaseBRIO, t)
	// CreateImage(CaseBRIO.bbox)
	// CaseBHAM := TestCase{[4]float64{10.000654, 52.517037, 13.38886, 53.550341}, 7, "Berlin - Hamburg"}
	// CheckCase(CaseBHAM, t)
	// CreateImage(CaseBHAM.bbox)

	// CaseBBARC := TestCase{[4]float64{-8.6107884, 41.1494512, 13.38886, 52.517037}, 4, "Berlin - Barcelona"}
	// CheckCase(CaseBBARC, t)
	// CaseBR := TestCase{[4]float64{12.482932, 41.89332, 13.38886, 52.517037}, 5, "Berlin - Rome"}
	// CheckCase(CaseBR, t)
	// bbox = min Longitude , min Latitude , max Longitude , max Latitude
	// CaseFlightFFM := TestCase{[4]float64{8.43103, 50.17878, 10.93463, 50.61335}, 7, "Flight around Frankfurt am Main"}
	// CheckCase(CaseFlightFFM, t)

}
