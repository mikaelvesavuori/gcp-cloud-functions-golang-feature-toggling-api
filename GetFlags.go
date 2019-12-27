// Package p contains an HTTP Cloud Function.
package p

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// RequestBody defines the required shape for POST requests
type RequestBody struct {
	Market string
}

// Flags struct which contains an array of feature flags
type Flags struct {
	Flags []FeatureFlag `json:"featureFlags"` // This maps to the top-level object name
}

// FeatureFlag defines the structure of the flags necessary per market
type FeatureFlag struct {
	Market            string           `json:"market"`
	NewFeatureActive  bool             `json:"newFeatureActive"`
	AbSplitPercentage *SplitPercentage `json:"abSplitPercentage,omitempty"`
}

// SplitPercentage specifies the splitting levels between new and current implementations
type SplitPercentage struct {
	New     int `json:"new"`
	Current int `json:"current"`
}

// GetFlags fetches feature flags
func GetFlags(w http.ResponseWriter, r *http.Request) {
	// Set initial, shared headers
	w.Header().Set("Access-Control-Allow-Origin", os.Getenv("ACCESS_CONTROL_ALLOW_ORIGIN"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	requestBody := RequestBody{}

	// Check for body
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		panic(err)
	}

	// Check that we don't receive empty/null market
	if requestBody.Market == "" {
		var notFoundErrorMessage = []byte(`Did not find any match`)
		fmt.Println(`Market is empty...?`)
		EndWithBadRequest(w, notFoundErrorMessage)
	}

	flags := GetBucketFlagData()

	var requestedMarket string = requestBody.Market

	var searchIndex int = FindMatch(flags, requestedMarket)

	// Return bad request if we didn't find a match
	if searchIndex == -1 {
		var notFoundErrorMessage = []byte(`Did not find any match`)
		fmt.Println(`Did not find any match`)
		EndWithBadRequest(w, notFoundErrorMessage)
		return
	}

	// Return data if we found a match
	if searchIndex != -1 {
		// Encode the data for delivery
		data, err := json.Marshal(flags.Flags[searchIndex])
		if err != nil {
			panic(err)
		}

		EndWithOK(w, data)
		return
	}
}

// GetBucketFlagData reads flags from a JSON file in a Google Cloud Storage bucket
func GetBucketFlagData() Flags {
	ctx := context.Background()

	// Set references
	bucketName := os.Getenv("BUCKET_NAME")
	bucketObject := os.Getenv("DATA_FILENAME")

	// Create GCS client
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Connect to storage
	bucket := client.Bucket(bucketName)
	query := &storage.Query{}
	it := bucket.Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		log.Println(attrs.Name)
	}

	// Read file
	obj := bucket.Object(bucketObject).ReadCompressed(false)
	rdr, err := obj.NewReader(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer rdr.Close()

	byteValue, _ := ioutil.ReadAll(rdr)

	// Decode multiple objects to Flags
	var flags Flags
	json.Unmarshal([]byte(byteValue), &flags)

	return flags
}

// FindMatch takes flag data and a requested market and will attempt to see if we have that market's data
func FindMatch(flags Flags, requestedMarket string) int {
	// Get the requested market index through iterating on the array
	var searchIndex int = -1

	for i := range flags.Flags {
		fmt.Println(flags.Flags[i])

		if flags.Flags[i].Market == requestedMarket {
			searchIndex = i
			break
		}
	}

	return searchIndex
}

// EndWithBadRequest handles finishing function calls that are not valid
func EndWithBadRequest(w http.ResponseWriter, err []byte) {
	w.Header().Set("Content-Type", "text/plain")
	http.Error(w, "Bad request", http.StatusBadRequest)
	return
}

// EndWithOK handles finishing function calls that are valid
func EndWithOK(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}
