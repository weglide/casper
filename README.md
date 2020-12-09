# Casper - static map generator for dynamic data

Casper generates static map images from hybrid map sources (e.g. raster tiles and mapbox vector tiles) and geojson data from postgres.

## Functionality flow idea

1. Grab the geojson data
2. Calculate needed tiles (xyz coordinates)
3. Merge tiles together to form map background
4. Overlay geojson
5. Serve as .jpg file

## AWS lambda extension

1. Use provided data form lambda function call instead of connecting directly to postgres.
5. Store .jpg file on AWS S3 and return path instead of serving as .jpg directly.

## Examples that provide similar experience

* Card style elements for tours on komoot.de
* Card style elements for tours on strava.com (login required)
* Mapbox static maps api with overlays 
* Google, Yandex etc. static map services
