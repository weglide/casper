package main

import (
	_ "image"
	_ "image/png"
	_ "log"
	"testing"
)

type TestCase struct {
	bbox      [4]float64
	ZoomLevel int16
	Name      string
}

func TestCaseBerlinNewYork(t *testing.T) {

	// BBox consist out of coordinates from Berlin and New York
	// Coordinates based on https://www.gps-coordinates.net/
	// Setup of different test cases to find zoom level and to Plot
	// Berlin - New York
	CaseBNY := TestCase{[4]float64{-74.006015, 40.71272, 13.38886, 52.517037}, 2, "Berlin - New York"}

	ImageBNY := NewImage(CaseBNY.bbox)

	// Find Tiles including the zoom level
	ImageBNY.FindRootTile()
	if ImageBNY.RootTile.X != 0 {
		t.Errorf("Roottile X is not equal to 0, Current Value :%d", ImageBNY.RootTile.X)
	}
	if ImageBNY.RootTile.Y != 0 {
		t.Errorf("Roottile Y is not equal to 0")
	}
	if ImageBNY.RootTile.Z != 0 {
		t.Errorf("Zoom Level Z is not equal to 0")
	}
	tiles, ZoomIncrease := TilesDownload(ImageBNY.RootTile.X, ImageBNY.RootTile.Y, ImageBNY.RootTile.Z)

	// Download Tiles with Zoom Level
	DownloadTiles(tiles, ImageBNY.RootTile.Z+ZoomIncrease)
	CreateImage(tiles, "BerlinNewYork")
	ImageBNY.DrawImage(&CaseBNY.bbox, tiles, ImageBNY.RootTile.Z, "BerlinNewYork", ImageBNY.RootTile.X, ImageBNY.RootTile.Y)

	// Check Image Berlin New York
	CheckImages("BerlinNewYork_merged_painted")
}

func TestCaseBerlinRio(t *testing.T) {
	// Berlin - Rio Case
	CaseBRIO := TestCase{[4]float64{-43.209373, -22.911014, 13.38886, 52.517037}, 2, "Berlin - RIO"}
	ImageBRIO := NewImage(CaseBRIO.bbox)

	// Find Tiles including the zoom level
	ImageBRIO.FindRootTile()
	ImageBRIO.ComposeImage("BerlinRio")
	tiles, _ := TilesDownload(ImageBRIO.RootTile.X, ImageBRIO.RootTile.Y, ImageBRIO.RootTile.Z)
	CreateImage(tiles, "BerlinRio")
	ImageBRIO.DrawImage(&CaseBRIO.bbox, tiles, ImageBRIO.RootTile.Z, "BerlinRio", ImageBRIO.RootTile.X, ImageBRIO.RootTile.Y)
}

func TestCaseBerlinHamburg(t *testing.T) {

	CaseBHAM := TestCase{[4]float64{10.000654, 52.517037, 13.38886, 53.550341}, 7, "Berlin - Hamburg"}
	ImageBHAM := NewImage(CaseBHAM.bbox)
	ImageBHAM.FindRootTile()
	if ImageBHAM.RootTile.X != 8 {
		t.Errorf("Roottile X is not equal to 0, the current values is %d", ImageBHAM.RootTile.X)
	}
	if ImageBHAM.RootTile.Y != 5 {
		t.Errorf("Roottile Y is not equal to 0, the current values is %d", ImageBHAM.RootTile.Y)
	}
	if ImageBHAM.RootTile.Z != 4 {
		t.Errorf("Zoom Level Z is not equal to 0, the current values is %d", ImageBHAM.RootTile.Z)
	}
	tiles, ZoomIncrease := TilesDownload(ImageBHAM.RootTile.X, ImageBHAM.RootTile.Y, ImageBHAM.RootTile.Z)

	// Download Tiles with Zoom Level
	DownloadTiles(tiles, ImageBHAM.RootTile.Z+ZoomIncrease)
	CreateImage(tiles, "BerlinNewYork")

	CreateImage(tiles, "BerlinHAM")
	ImageBHAM.DrawImage(&CaseBHAM.bbox, tiles, ImageBHAM.RootTile.Z, "BerlinHAM", ImageBHAM.RootTile.X, ImageBHAM.RootTile.Y)

	CheckImages("BerlinHAM_merged_painted")
}

func TestCaseBerlinBarcelona(t *testing.T) {
	CaseBBARC := TestCase{[4]float64{2.154007, 41.390205, 13.38886, 52.517037}, 4, "Berlin - Barcelona"}
	ImageBBARC := NewImage(CaseBBARC.bbox)
	// Find Tiles including the zoom level
	ImageBBARC.FindRootTile()
	tiles, ZoomIncrease := TilesDownload(ImageBBARC.RootTile.X, ImageBBARC.RootTile.Y, ImageBBARC.RootTile.Z)
	// Download Tiles with Zoom Level
	DownloadTiles(tiles, ImageBBARC.RootTile.Z+ZoomIncrease)
	// ImageBBARC.ComposeImage("BerlinBBARC")
	CreateImage(tiles, "BerlinBBARC")
	ImageBBARC.DrawImage(&ImageBBARC.bbox, tiles, ImageBBARC.RootTile.Z, "BerlinBBARC", ImageBBARC.RootTile.X, ImageBBARC.RootTile.Y)
	CheckImages("BerlinBBARC_merged_painted")
}

func TestFindRootTile(t *testing.T) {

	// bbox = min Longitude , min Latitude , max Longitude , max Latitude
	CaseFlightFFM := TestCase{[4]float64{8.682127, 50.110922, 8.7667933, 50.8021728}, 7, "Flight around Frankfurt am Main"}
	ImageFlightFFM := NewImage(CaseFlightFFM.bbox)
	// Find Tiles including the zoom level
	ImageFlightFFM.FindRootTile()
	ImageFlightFFM.ComposeImage("FlightFFM")
	// CheckImages("FlightFFM_merged")
	tiles, ZoomIncrease := TilesDownload(ImageFlightFFM.RootTile.X, ImageFlightFFM.RootTile.Y, ImageFlightFFM.RootTile.Z)
	DownloadTiles(tiles, ImageFlightFFM.RootTile.Z+ZoomIncrease)
	CreateImage(tiles, "FlightFFM")
	ImageFlightFFM.DrawImage(&ImageFlightFFM.bbox, tiles, ImageFlightFFM.RootTile.Z, "FlightFFM", ImageFlightFFM.RootTile.X, ImageFlightFFM.RootTile.Y)
	CheckImages("FlightFFM_merged_painted")

}
