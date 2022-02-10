package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	waitAfterAdd = time.Millisecond * 100
)

var (
	re = regexp.MustCompile(`'({[^']+})'`)
)

type simulationConfig struct {
	DataDir            string
	AnonymityThreshold int
	CrowdIDMethod      int
	CSVOutput          bool
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

	var ms []Report
	for _, line := range strings.Split(string(content), "\n") {
		// This is (probably) the file's trailing newline but we continue, just
		// in case.
		if line == "" {
			continue
		}

		// Extract P3A measurement from the current line.
		measurements := re.FindStringSubmatch(line)
		if len(measurements) == 0 {
			continue
		}
		if len(measurements) != 2 {
			return nil, fmt.Errorf("line does not contain exactly one measurement: %s", line)
		}

		var m P3AMeasurement
		buf := bytes.NewBufferString(measurements[1])
		if err = json.NewDecoder(buf).Decode(&m); err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	return ms, nil
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
		if err != nil {
			return err
		}
		shufflerInbox <- rs
	}
	return nil
}

func simulationMode(cfg *simulationConfig) {
	s := NewShuffler(batchPeriod, cfg.AnonymityThreshold, cfg.CrowdIDMethod)
	s.Start()

	if err := parseDir(cfg.DataDir, s.inbox); err != nil {
		log.Fatalf("Failed to load P3A reports from directory: %s", err)
	}
	// Give the shuffler a little bit of time to add pending reports before we
	// proceed.  It's not pretty but it will do for now.
	time.Sleep(waitAfterAdd)
	log.Printf("Simulate: Before batch period: %s\n", s)

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
	// Same as above.
	time.Sleep(waitAfterAdd)
	log.Printf("Simulate: After batch period: %s\n", s)
	if cfg.CSVOutput {
		fmt.Printf("%d,%d,%d,%d\n", cfg.CrowdIDMethod, cfg.AnonymityThreshold, s.briefcase.NumCrowdIDs(), s.briefcase.NumReports())
	}
}
