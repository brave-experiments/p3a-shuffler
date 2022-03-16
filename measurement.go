package main

import (
	"crypto/sha1"
	"fmt"
	"strings"
)

const (
	attrsAll     = iota // All attributes are used for k-anonymity.
	attrsNoValue        // All attributes *except* the measurement value.
	attrsMinimal        // An ideal subset of attributes.
)

var (
	anonymityAttrs = map[int]string{
		attrsAll:     "All",
		attrsNoValue: "NoValue",
		attrsMinimal: "Minimal",
	}
)

// ShufflerMeasurement represents an encrypted measurement for the shuffler.
type ShufflerMeasurement struct {
	Encrypted []byte `json:"encrypted"`
}

// P3AMeasurement represents a P3A measurement as it's sent by Brave clients.
// See the browser code for how measurements are created:
// https://github.com/brave/brave-core/blob/1adaa0bc057a83f432e9c278c7c373ef60a5b766/components/p3a/p3a_measurement.cc#L70
// P3AMeasurement also implements the Report interface.
type P3AMeasurement struct {
	YearOfSurvey  int    `json:"yos"`
	YearOfInstall int    `json:"yoi"`
	WeekOfSurvey  int    `json:"wos"`
	WeekOfInstall int    `json:"woi"`
	MetricValue   int    `json:"metric_value"`
	MetricName    string `json:"metric_name"`
	CountryCode   string `json:"country_code"`
	Platform      string `json:"platform"`
	Version       string `json:"version"`
	Channel       string `json:"channel"`
	RefCode       string `json:"refcode"`
}

// IsValid returns true if the given P3A measurement is valid.
func (m P3AMeasurement) IsValid() bool {
	if m.YearOfSurvey < 1970 || m.YearOfInstall < 1970 {
		return false
	}
	if m.WeekOfSurvey < 1 || m.WeekOfSurvey > 53 {
		return false
	}
	if m.WeekOfInstall < 1 || m.WeekOfInstall > 53 {
		return false
	}
	if m.MetricValue < 0 {
		return false
	}
	if m.MetricName == "" {
		return false
	}
	if m.Platform == "" || m.Version == "" {
		return false
	}
	if m.Channel == "" {
		return false
	}
	return true
}

// String returns a human-readable string representation of the P3A
// measurement.
func (m P3AMeasurement) String() string {
	return fmt.Sprintf("P3A measurement:\n"+
		"\tYear of survey:  %d\n"+
		"\tYear of install: %d\n"+
		"\tWeek of survey:  %d\n"+
		"\tWeek of install: %d\n"+
		"\tMetric value:    %d\n"+
		"\tMetric name:     %s\n"+
		"\tCountry code:    %s\n"+
		"\tPlatform:        %s\n"+
		"\tVersion:         %s\n"+
		"\tChannel:         %s\n"+
		"\tRefcode:         %s\n",
		m.YearOfSurvey, m.YearOfInstall,
		m.WeekOfSurvey, m.WeekOfInstall,
		m.MetricValue, m.MetricName,
		m.CountryCode, m.Platform, m.Version,
		m.Channel, m.RefCode)
}

// OrderHighEntropyFirst turns the measurement into a a slice of strings,
// ordered by entropy, with high-entropy attributes coming first.  The argument
// 'method' determines what attributes are returned.
func (m P3AMeasurement) OrderHighEntropyFirst(method int) []string {
	switch method {
	case attrsAll:
		return []string{
			m.MetricName,                       // 0.93
			fmt.Sprintf("%d", m.MetricValue),   // 0.64
			fmt.Sprintf("%d", m.WeekOfInstall), // 0.86
			m.CountryCode,                      // 0.70
			m.Version,                          // 0.66
			m.RefCode,                          // 0.62
			m.Platform,                         // 0.61
			m.Channel,                          // 0.55
			fmt.Sprintf("%d", m.YearOfInstall), // 0.42
			fmt.Sprintf("%d", m.WeekOfSurvey),  // 0.06
			fmt.Sprintf("%d", m.YearOfSurvey),  // 0.00
		}
	case attrsNoValue:
		return []string{
			m.MetricName,                       // 0.93 (normalized entropy)
			fmt.Sprintf("%d", m.WeekOfInstall), // 0.86
			m.CountryCode,                      // 0.70
			m.Version,                          // 0.66
			m.RefCode,                          // 0.62
			m.Platform,                         // 0.61
			m.Channel,                          // 0.55
			fmt.Sprintf("%d", m.YearOfInstall), // 0.42
			fmt.Sprintf("%d", m.WeekOfSurvey),  // 0.06
			fmt.Sprintf("%d", m.YearOfSurvey),  // 0.00
		}
	case attrsMinimal:
		return []string{
			m.MetricName,                       // 0.93
			m.Channel,                          // 0.55
			m.Platform,                         // 0.61
			m.CountryCode,                      // 0.70
			fmt.Sprintf("%d", m.WeekOfInstall), // 0.86
		}
	default:
		elog.Fatalf("Unexpected method for measurement: %d", method)
		return []string{}
	}
}

// OrderHighEntropyLast returns the reverse ordering of OrderHighEntropyFirst.
func (m P3AMeasurement) OrderHighEntropyLast(method int) []string {
	orig := m.OrderHighEntropyFirst(method)
	reversed := []string{}
	for i := len(orig) - 1; i >= 0; i-- {
		reversed = append(reversed, orig[i])
	}
	return reversed
}

// CSVHeader returns the header for a CSV-formatted file that contains P3A
// measurements.
func (m P3AMeasurement) CSVHeader() string {
	return "yos,yoi,wos,woi,metric_value,metric_name,country_code,platform,version,channel,refcode"
}

// CSV returns the measurement with all its attributes as CSV record.
func (m P3AMeasurement) CSV() string {
	return fmt.Sprintf("%d,%d,%d,%d,%d,%s,%s,%s,%s,%s,%s",
		m.YearOfSurvey, m.YearOfInstall,
		m.WeekOfSurvey, m.WeekOfInstall,
		m.MetricValue, m.MetricName,
		m.CountryCode, m.Platform, m.Version,
		m.Channel, m.RefCode)
}

// CrowdID returns the crowd ID of the P3A measurement.
func (m P3AMeasurement) CrowdID(method int) CrowdID {
	// No need to do anything extravagant here.  Simply get an ordered slice
	// for the given measurement, hash it, and return it.
	attrs := m.OrderHighEntropyFirst(method)
	payload := []byte(strings.Join(attrs, ""))
	hash := fmt.Sprintf("%x", sha1.Sum(payload))
	return CrowdID(hash)
}

// Payload returns the P3A measurement's payload.
func (m P3AMeasurement) Payload() []byte {
	return []byte(m.String())
}
