package main

import (
	"log"
	"os"
	"testing"
	"time"
)

func BenchmarkRealWorldData(b *testing.B) {
	// Set this environment variable to a local directory containing P3A data
	// (as it can be found in the S3 bucket) to run the benchmarks.
	p3aDataDir := os.Getenv("P3A_DATA_DIR")
	if p3aDataDir == "" {
		return
	}

	s := NewShuffler(time.Hour, anonymityThreshold, defaultCrowdIDMethod)
	s.Start()

	for n := 0; n < b.N; n++ {
		reports, err := parseReportsFromDir(p3aDataDir)
		if err != nil {
			b.Fatalf("Failed to load P3A reports from directory: %s", err)
		}
		s.inbox <- reports

		go func() {
			for {
				<-s.outbox
			}
		}()
		if err := s.endBatchPeriod(); err != nil {
			log.Fatal(err)
		}
	}
}
