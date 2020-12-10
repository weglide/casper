#!/usr/bin/env bash
# Simple shell script to install dependencies
packages=(
    'github.com/paulmach/orb'
    'github.com/gorilla/mux'
    'github.com/gogo/protobuf/proto'
    'github.com/pkg/errors'
    'github.com/lib/pq'
    'github.com/githubnemo/CompileDaemon'
    'github.com/aws/aws-lambda-go/lambda'
    'github.com/aws/aws-lambda-go/events'
)
for package in ${packages[@]}; do
    echo "go get ${package}"
    go get ${package}
done

