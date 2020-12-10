# Casper - static map generator for dynamic data

Casper generates static map images from hybrid map sources (e.g. raster tiles and mapbox vector tiles) and geojson data from postgres.

## Functionality flow idea

1. Grab the geojson data
2. Calculate needed tiles (xyz coordinates)
3. Merge tiles together to form map background
4. Overlay geojson
5. Serve as .jpg file

## Local Development

1. Set Local environment variable with `export LOCAL=True`
2. Build go executable `go build main.go`
3. Run executable `./main`

Instead of using the commands you can build and and run the executable with the shell script `run.sh`.

## Data input format

Input will be LineString of length < 1000 (ST_Simplify) in the backend before -> more points will not be visible on the static map. Data input could be a geojson file or WKT representation of geometry, what is more sensible? Geojson probably more generic?

## Generic

Service should be able to iterate a list of tile endpoints (Airspace, Elevation styled) and render on image. Specify inputs in envs.

## AWS lambda extension

1. CI to deploy function to AWS Lambda, examples for Go & Rust in other repositories.
2. Deployment package size needs to be below 50MB zipped -> python with scipy & numpy & PIL could be problematic
3. Store .jpg output file on AWS S3 and return path instead of serving as .jpg directly.

## Examples that provide similar experience

* Card style elements for tours on komoot.de
* Card style elements for tours on strava.com (login required)
* Mapbox static maps api with overlays 
* Google, Yandex etc. static map services
* bullet point
