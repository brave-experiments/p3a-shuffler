package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	re = regexp.MustCompile(`'({[^']+})'`)
)

type simulationConfig struct {
	DataDir            string
	AnonymityThreshold int
	CrowdIDMethod      int
	Order              int
	AttributeCSV       bool
	Entropy            bool
}

// parseJSONFile reads and parses a P3A measurement file as it can be found in
// our S3 bucket.  Those files have the following (sanitized) format:
//
// <134>2022-01-01T00:00:00Z foo bar[quuz]: "-" "-" 2022-01-01:00:xx:xx
// POST / HTTP/2 200 '{"channel":"nightly","country_code":"US","metric_name":
// "...","metric_value":0,"platform":"linux-bc","refcode":"none",
// "version":"1.36.46","woi":3,"wos":3,"yoi":2022,"yos":2022}'
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
		if !m.IsValid() {
			continue
		}
		ms = append(ms, m)
	}
	return ms, nil
}

// empiricalEntropyByField determines the empirical entropy per measurement
// attribute.
func empiricalEntropyByField(rs []Report) {
	yos := make(map[string]int)
	yoi := make(map[string]int)
	wos := make(map[string]int)
	woi := make(map[string]int)
	metricValue := make(map[string]int)
	metricName := make(map[string]int)
	countryCode := make(map[string]int)
	platform := make(map[string]int)
	version := make(map[string]int)
	channel := make(map[string]int)
	refcode := make(map[string]int)

	for _, r := range rs {
		m := r.(P3AMeasurement)
		incKey(fmt.Sprintf("%d", m.YearOfInstall), yoi)
		incKey(fmt.Sprintf("%d", m.YearOfSurvey), yos)
		incKey(fmt.Sprintf("%d", m.WeekOfInstall), woi)
		incKey(fmt.Sprintf("%d", m.WeekOfSurvey), wos)
		incKey(fmt.Sprintf("%d", m.MetricValue), metricValue)
		incKey(m.MetricName, metricName)
		incKey(m.CountryCode, countryCode)
		incKey(m.Platform, platform)
		incKey(m.Version, version)
		incKey(m.Channel, channel)
		incKey(m.RefCode, refcode)

	}
	fmt.Printf("Entropy for yoi: %.2f\n", empiricalEntropy(yoi))
	fmt.Printf("Entropy for yos: %.2f\n", empiricalEntropy(yos))
	fmt.Printf("Entropy for woi: %.2f\n", empiricalEntropy(woi))
	fmt.Printf("Entropy for wos: %.2f\n", empiricalEntropy(wos))
	fmt.Printf("Entropy for metric_value: %.2f\n", empiricalEntropy(metricValue))
	fmt.Printf("Entropy for metric_name: %.2f\n", empiricalEntropy(metricName))
	fmt.Printf("Entropy for country_code: %.2f\n", empiricalEntropy(countryCode))
	fmt.Printf("Entropy for platform: %.2f\n", empiricalEntropy(platform))
	fmt.Printf("Entropy for version: %.2f\n", empiricalEntropy(version))
	fmt.Printf("Entropy for channel: %.2f\n", empiricalEntropy(channel))
	fmt.Printf("Entropy for refcode: %.2f\n", empiricalEntropy(refcode))
}

func empiricalEntropy(m map[string]int) float64 {
	// Get the total number of elements.
	total := 0
	for _, num := range m {
		total += num
	}

	// Calculate the empirical entropy.
	var entropy float64
	for _, num := range m {
		p := float64(num) / float64(total)
		entropy += p * math.Log2(p)
	}

	l := math.Log2(float64(len(m)))
	if l == 0 {
		return float64(0)
	}
	// Normalize entropy to [0, 1].
	return -entropy / l
}

func incKey(key string, m map[string]int) {
	v, exists := m[key]
	if !exists {
		m[key] = 1
	} else {
		m[key] = v + 1
	}
}

func simulateShuffler(cfg *simulationConfig, reports []Report) {
	var origReports int

	s := NewShuffler(batchPeriod, cfg.AnonymityThreshold, cfg.CrowdIDMethod)
	s.Start()
	s.inbox <- reports

	elog.Printf("Before batch period: %s\n", s)
	origReports = s.briefcase.NumReports()

	elog.Printf("Ending batch period using anonymity threshold of %d.", cfg.AnonymityThreshold)
	s.briefcase.DumpFewerThan(s.anonymityThreshold)
	elog.Printf("After batch period: %s\n", s)

	fmt.Printf("%s,%d,%d,%.3f,0,0,0,0\n",
		anonymityAttrs[cfg.CrowdIDMethod],
		cfg.Order,
		cfg.AnonymityThreshold,
		frac(s.briefcase.NumReports(),
			origReports))
}

func simulateSTAR(cfg *simulationConfig, reports []Report) {
	s := NewNestedSTAR(cfg)

	numAttrs := len(P3AMeasurement{}.OrderHighEntropyFirst(cfg.CrowdIDMethod))

	s.AddReports(cfg.CrowdIDMethod, reports)
	elog.Printf("Aggregating %d measurements using k=%d, method=%d, attrs=%d.",
		s.numMeasurements, cfg.AnonymityThreshold, cfg.CrowdIDMethod, numAttrs)
	s.Aggregate(cfg.CrowdIDMethod, numAttrs)
}

// parseReportsFromDir parses and returns all P3A measurements from the files
// that can be found in the given directory (and subdirectories).
func parseReportsFromDir(dir string) ([]Report, error) {
	var reports []Report
	var numFiles int
	defer func() {
		elog.Printf("Parsed %d JSON files.", numFiles)
	}()

	err := filepath.Walk(dir,
		func(filename string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			numFiles++
			rs, err := parseJSONFile(filename)
			if err != nil {
				elog.Printf("Failed to parse %s because: %s", filename, err)
			}
			reports = append(reports, rs...)
			return nil
		})
	if err != nil {
		return nil, err
	}
	return reports, nil
}

func attributeCSV(cfg *simulationConfig, reports []Report) {
	elog.Println("Printing per-attribute CSVs.")
	fmt.Println(P3AMeasurement{}.CSVHeader())
	for i, r := range reports {
		if i%1000 == 0 {
			elog.Printf("Processed %d measurements.", i)
		}
		fmt.Println(r.(P3AMeasurement).CSV())
	}
}

func simulationMode(cfg *simulationConfig) {
	elog.Printf("Starting to read reports from %s.", cfg.DataDir)
	reports, err := parseReportsFromDir(cfg.DataDir)
	if err != nil {
		elog.Fatalf("Failed to parse measurements: %s", err)
	}
	elog.Printf("Read %d P3A measurements from disk.", len(reports))

	if cfg.AttributeCSV {
		attributeCSV(cfg, reports)
		return
	}
	if cfg.Entropy {
		empiricalEntropyByField(reports)
		return
	}
	cfg.Order = orderHighEntropyLast

	fmt.Println("method,order,threshold,reports,num_tags,num_leaf_tags,len_part_msmts,num_part_msmts")

	// Iterate over our desired k-anonymity thresholds.
	thresholds := []int{5, 10, 20, 40, 80, 160, 320}
	for _, k := range thresholds {
		cfg.AnonymityThreshold = k

		for method, name := range anonymityAttrs {
			elog.Printf("Running simulation for k=%d, method=%s", k, name)
			cfg.CrowdIDMethod = method
			simulateShuffler(cfg, reports)
			simulateSTAR(cfg, reports)
		}
	}
}
