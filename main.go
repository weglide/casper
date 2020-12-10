package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/mvt"
	"github.com/paulmach/orb/encoding/wkb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/paulmach/orb/simplify"

	"github.com/lib/pq"

	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// required input params
type Params struct {
	ids  []int
	zoom maptile.Zoom
	x    uint32
	y    uint32
}

type Response struct {
	body            string
	isBase64Encoded bool
	statusCode      int
	headers         string
}

func main() {
	LOCAL, _ := strconv.ParseBool(os.Getenv("LOCAL"))

	// switch between lambda and local environment
	if LOCAL == true {
		r := mux.NewRouter()
		// e.g. localhost:6969/flights/12,13/8/133/86.pbf
		r.HandleFunc("/flights/{ids}/{z}/{x}/{y}", flightHandlerLocal)
		http.ListenAndServe(":6969", r)
	} else {
		lambda.Start(flightHandlerLambda)
	}
}

func flightHandlerLocal(w http.ResponseWriter, r *http.Request) {
	params, err := unmarshalParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tile := maptile.New(params.x, params.y, params.zoom)

	featureCollection, err := fetchFeatureCollection(params.ids, tile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	encoded, err := encodeMvt(featureCollection, tile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set response headers
	mvtResponseHeaders(len(encoded), w)

	// output
	w.Write(encoded)
}

// handle tile request
func flightHandlerLambda(ctx context.Context, r *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	params, err := unmarshalParams(r.PathParameters)

	if err != nil {
		log.Printf("Error: %s", err)
		return nil, err
	}

	tile := maptile.New(params.x, params.y, params.zoom)

	featureCollection, err := fetchFeatureCollection(params.ids, tile)

	if err != nil {
		log.Printf("Error: %s", err)

		return nil, err
	}

	encoded, err := encodeMvt(featureCollection, tile)
	if err != nil {
		log.Printf("Error: %s", err)

		return nil, err
	}

	// set response headers
	// output
	return &events.APIGatewayProxyResponse{
		Body:            EncodeB64(encoded),
		IsBase64Encoded: true,
		StatusCode:      200,
		Headers: map[string]string{
			"Content-Type":                "application/x-protobuf",
			"Access-Control-Allow-Origin": "*",
			"content-encoding":            "gzip",
			"content-length":              strconv.Itoa(len(encoded)),
		},
	}, nil
}

func EncodeB64(message []byte) (retour string) {
	base64Text := make([]byte, base64.StdEncoding.EncodedLen(len(message)))
	base64.StdEncoding.Encode(base64Text, []byte(message))
	return string(base64Text)
}

// unmarshal url parameters to struct (safely?)
func unmarshalParams(r interface{}) (Params, error) {
	p := make(map[string]string)

	if r2, ok := r.(*http.Request); ok {
		p = mux.Vars(r2)
	} else if r2, ok2 := r.(map[string]string); ok2 {
		p = r2
	}

	var ids []int
	// string of integer list to json array ("1,3,2" --> "[1,3,2]")
	jsonIds := []byte(fmt.Sprintf("[%s]", p["ids"]))

	// json --> array
	err := json.Unmarshal(jsonIds, &ids)

	// parse as uint32  (base=10, bit=32)
	z, err := strconv.ParseUint(p["z"], 10, 32)
	x, err := strconv.ParseUint(p["x"], 10, 32)
	y, err := strconv.ParseUint(p["y"], 10, 32)

	// build Params struct and return
	return Params{
		ids,
		maptile.Zoom(z),
		uint32(x),
		uint32(y),
	}, err
}

func psqlConnectionString() string {
	// get environment connection vars
	var (
		host     = os.Getenv("POSTGRES_HOST")
		port     = os.Getenv("POSTGRES_PORT")
		user     = os.Getenv("POSTGRES_USER")
		password = os.Getenv("POSTGRES_PASS")
		dbname   = os.Getenv("POSTGRES_DB")
	)

	// build connection string
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
}

// fetch line strings from db by ids
func fetchFeatureCollection(ids []int, tile maptile.Tile) (*geojson.FeatureCollection, error) {
	// open connection
	db, err := sql.Open("postgres", psqlConnectionString())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	bound := tile.Bound()

	// execute query
	rows, err := db.Query(
		"SELECT id, bbox, ST_AsBinary(line_wkt) FROM flight WHERE id = ANY($1) AND NOT (bbox[1] > $2 OR bbox[2] > $3 OR bbox[3] < $4 OR bbox[4] < $5)", 
		pq.Array(ids), bound.Max[0], bound.Max[1], bound.Min[0], bound.Min[1],
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	featureCollection := geojson.NewFeatureCollection()

	// loop trough rows and assemble geojson feature collection
	for rows.Next() {
		var id int
		var bbox string
		var line orb.LineString

		err = rows.Scan(&id, &bbox, wkb.Scanner(&line))
		if err != nil {
			return nil, err
		}

		feature := geojson.NewFeature(line)
		feature.Properties = geojson.Properties{
			"id": id,
		}

		featureCollection.Append(feature)
	}

	return featureCollection, nil
}

func encodeMvt(featureCollection *geojson.FeatureCollection, tile maptile.Tile) ([]byte, error) {
	// create mapbox vector tiles from geojson feature collection
	layers := mvt.NewLayers(map[string]*geojson.FeatureCollection{"flights": featureCollection})

	// set proper mvt version, see https://github.com/paulmach/orb/issues/11
	for i, _ := range layers {
		layers[i].Version = 2
	}

	// project to the tile coords
	layers.ProjectToTile(tile)

	// In order to be used as source for MapboxGL geometries need to be clipped
	// to max allowed extent. (uncomment next line)
	layers.Clip(mvt.MapboxGLDefaultExtentBound)

	// Simplify the geometry now that it's in tile coordinate space.
	layers.Simplify(simplify.DouglasPeucker(3.0))

	// Depending on use-case remove empty geometry, those too small to be
	// represented in this tile space.
	// In this case lines shorter than 1, and areas smaller than 2.
	// layers.RemoveEmpty(1.0, 2.0)

	// marshal and gzip to byte array
	return mvt.MarshalGzipped(layers)
}

// set response headers for gzipped mapbox vector tile
func mvtResponseHeaders(size int, w http.ResponseWriter) {
	w.Header().Set("content-encoding", "gzip")
	w.Header().Set("content-type", "application/x-protobuf")
	w.Header().Set("access-control-allow-origin", "*")
	w.Header().Set("content-length", strconv.Itoa(size))
}
