package main

// This file takes reports from the shuffler's outbox and forwards them to an
// external machine.

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

// Forwarder is responsible for forwarding shuffled reports to the server
// (a.k.a. the analyzer in PROCHLO).
type Forwarder struct {
	sync.WaitGroup
	done     chan bool
	srvURL   string
	shuffler chan []Report
}

// NewForwarder creates and returns a new forwarder.
func NewForwarder(shuffler chan []Report, srvURL string) *Forwarder {
	return &Forwarder{
		done:     make(chan bool),
		shuffler: shuffler,
		srvURL:   srvURL,
	}
}

// Start starts the forwarder.
func (f *Forwarder) Start() {
	f.Add(1)
	go func() {
		defer f.Done()
		for {
			select {
			case <-f.done:
				return
			case reports := <-f.shuffler:
				log.Printf("Forwarder: Received %d reports from shuffler.", len(reports))
				go f.forward(reports)
			}
		}
	}()
}

// Stop stops the forwarder.
func (f *Forwarder) Stop() {
	f.done <- true
	f.Wait()
}

// forward forwards the given reports to the server.
func (f *Forwarder) forward(reports []Report) {
	if len(reports) == 0 {
		log.Println("Forwarder: No reports given, so there's nothing to forward.")
		return
	}

	// Marshal our reports.
	type reportBatch struct {
		Batch []Report `json:"batch"`
	}
	batch := reportBatch{Batch: reports}
	jsonBytes, err := json.Marshal(batch)
	if err != nil {
		log.Printf("Forwarder: Failed to marshal reports: %s", err)
		return
	}

	// Forward our JSON blob to the server.
	// TODO: We should probably try to re-submit the reports in case of error.
	// Otherwise, we would lose reports.
	r := bytes.NewReader(jsonBytes)
	resp, err := http.Post(f.srvURL, "application/json", r)
	if err != nil {
		log.Printf("Forwarder: Failed to POST reports to server: %s", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("Forwarder: Received HTTP status code %d from server.", resp.StatusCode)
		return
	}

	log.Printf("Forwarder: Forwarded %d reports to server.", len(reports))
}
