package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"sync"
	"time"
)

var (
	re           = regexp.MustCompile(`'({[^']+})'`)
	errEmptyFile = errors.New("file contains no P3A measurement")
)

type simulationConfig struct {
	DataDir            string
	AnonymityThreshold int
}

// parseJSONFile reads and parses a P3A measurement file as it can be found in
// our S3 bucket.  Those files have the following (sanitized) format:
//
// <134>2022-01-01T00:00:00Z foo bar[quuz]: "-" "-" 2022-01-01:00:xx:xx
// POST / HTTP/2 200 '{"channel":"nightly","country_code":"US","metric_hash":
// "0000000000000000000","metric_value":0,"platform":"linux-bc","refcode":
// "none","version":"1.36.46","woi":3,"wos":3,"yoi":2022,"yos":2022}'
func parseJSONFile(filename string) ([]Report, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Extract JSON string from blob.
	measurements := re.FindStringSubmatch(string(content))
	if len(measurements) == 0 {
		return nil, errEmptyFile
	}
	if len(measurements) != 2 {
		return nil, errors.New("unexpected number of measurements")
	}

	var m P3AMeasurement
	buf := bytes.NewBufferString(measurements[1])
	if err = json.NewDecoder(buf).Decode(&m); err != nil {
		return nil, err
	}
	return []Report{m}, nil
}

// parseDir parses all P3A measurements from the files that can be found in the
// given directory and sends the measurements to the shuffler's inbox.
func parseDir(dir string, shufflerInbox chan []Report) error {
	fileInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range fileInfo {
		rs, err := parseJSONFile(path.Join(dir, file.Name()))
		if err == errEmptyFile {
			continue
		}
		if err != nil {
			return err
		}
		shufflerInbox <- rs
	}
	return nil
}

func simulationMode(cfg *simulationConfig) {
	s := NewShuffler(batchPeriod, cfg.AnonymityThreshold)
	s.Start()

	if err := parseDir(cfg.DataDir, s.inbox); err != nil {
		log.Fatalf("Failed to load P3A reports from directory: %s", err)
	}
	log.Printf("Simulate: %s", s)

	var rs []Report
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			rs = append(rs, <-s.outbox...)
			wg.Done()
		}
	}()
	log.Printf("Simulate: Ending batch period using anonymity threshold of %d.", cfg.AnonymityThreshold)
	if err := s.endBatchPeriod(); err != nil {
		log.Fatal(err)
	}
	wg.Wait()

	s.inbox <- rs
	// Give the shuffler a second to add the reports to the briefcase.
	time.Sleep(time.Second)
	log.Printf("Simulate: %s", s)
}
