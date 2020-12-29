package main

import (
	"log"
	"testing"
)

func TestNum2deg(t *testing.T) {
	testTile := new(Tile)
	testTile.X = 10
	testTile.Z = 10
	log.Println(testTile.Num2deg())
}
