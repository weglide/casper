#!/usr/bin/env bash
# Simple shell script to install dependencies
packages=(
    "github.com/fogleman/gg"
	"github.com/lib/pq"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkb"
	"github.com/paulmach/orb/geojson"
	"github.com/oliamb/cutter"
)
for package in ${packages[@]}; do
    echo "go get ${package}"
    go get ${package}
done

