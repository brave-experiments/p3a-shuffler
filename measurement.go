package main

import (
	"crypto/sha1"
	"fmt"
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
	MetricHash    string `json:"metric_hash"`
	CountryCode   string `json:"country_code"`
	Platform      string `json:"platform"`
	Version       string `json:"version"`
	Channel       string `json:"channel"`
	RefCode       string `json:"refcode"`
}

// String returns a string representation of the P3A measurement.
func (m P3AMeasurement) String() string {
	return fmt.Sprintf("P3A measurement:\n"+
		"\tYear of survey:  %d\n"+
		"\tYear of install: %d\n"+
		"\tWeek of survey:  %d\n"+
		"\tWeek of install: %d\n"+
		"\tMetric value:    %d\n"+
		"\tMetric hash:     %s\n"+
		"\tCountry code:    %s\n"+
		"\tPlatform:        %s\n"+
		"\tVersion:         %s\n"+
		"\tChannel:         %s\n"+
		"\tRefcode:         %s\n",
		m.YearOfSurvey, m.YearOfInstall,
		m.WeekOfSurvey, m.WeekOfInstall,
		m.MetricValue, m.MetricHash,
		m.CountryCode, m.Platform, m.Version,
		m.Channel, m.RefCode)
}

// CrowdID returns the crowd ID (a SHA-1 over the measurement) of the P3A
// measurement.
func (m P3AMeasurement) CrowdID() CrowdID {
	hash := fmt.Sprintf("%x", sha1.Sum(m.Payload()))
	return CrowdID(hash)
}

// Payload returns the P3A measurement's payload.
func (m P3AMeasurement) Payload() []byte {
	return []byte(m.String())
}
