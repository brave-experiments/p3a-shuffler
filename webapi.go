package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// createP3AHandler creates a handler that receives a set of JSON-encoded P3A
// measurements.
func createP3AHandler(inbox chan []Report) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ms []P3AMeasurement

		err := json.NewDecoder(r.Body).Decode(&ms)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		rs := []Report{}
		for _, m := range ms {
			rs = append(rs, m)
		}

		inbox <- rs
		log.Printf("WebAPI: Sent %d P3A measurement to shuffler.", len(ms))
	}
}

// createShufflerHandler creates a handler that receives an encrypted blob
// that, when encrypted, contains a JSON-encoded structure consisting of a
// crowd ID and an encrypted payload that is opaque to the shuffler.
func createShufflerHandler(inbox chan []Report) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Decrypt report and forward it to the shuffler's inbox.
	}
}
