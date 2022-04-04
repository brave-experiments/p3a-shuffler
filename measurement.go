package main

import (
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"
)

const (
	attrsAll        = iota // All attributes are used for k-anonymity.
	attrsRefactored        // Our current set of attributes.
	attrsMinimal           // A minimal set of attributes.
)

var (
	anonymityAttrs = map[int]string{
		attrsAll:        "All",
		attrsRefactored: "Refactored",
		attrsMinimal:    "Minimal",
	}
	lastVersion = map[string]*version{
		"nightly":   newVersion("0.0.0"),
		"release":   newVersion("0.0.0"),
		"beta":      newVersion("0.0.0"),
		"canary":    newVersion("0.0.0"),
		"dev":       newVersion("0.0.0"),
		"developer": newVersion("0.0.0"),
		"unknown":   newVersion("0.0.0"),
		"":          newVersion("0.0.0"),
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
			// The following entropy numbers came from running:
			// p3a-shuffler -simulate -datadir 2022-03-27 -entropy
			m.MetricName,                       // 0.90
			fmt.Sprintf("%d", m.MetricValue),   // 0.66
			fmt.Sprintf("%d", m.WeekOfInstall), // 0.93
			m.CountryCode,                      // 0.72
			m.Platform,                         // 0.57
			fmt.Sprintf("%d", m.YearOfInstall), // 0.40
			m.Version,                          // 0.25
			m.RefCode,                          // 0.17
			fmt.Sprintf("%d", m.WeekOfSurvey),  // 0.15
			m.Channel,                          // 0.03
			fmt.Sprintf("%d", m.YearOfSurvey),  // 0.00
		}
	case attrsMinimal:
		return []string{
			m.MetricName,
			fmt.Sprintf("%d", m.MetricValue),
			fmt.Sprintf("%d", m.WeekOfInstall),
			m.CountryCode,
			m.Platform,
			m.Channel,
			fmt.Sprintf("%t", isRecentVersion(m.Channel, m.Version)),
		}
	case attrsRefactored:
		return []string{
			m.MetricName,
			fmt.Sprintf("%d", m.MetricValue),
			fmt.Sprintf("%d", m.WeekOfInstall),
			m.CountryCode,
			m.Platform,
			m.Channel,
			fmt.Sprintf("%d", m.YearOfInstall),
			fmt.Sprintf("%d", m.WeekOfSurvey),
			fmt.Sprintf("%d", m.YearOfSurvey),
			fmt.Sprintf("%t", isRecentVersion(m.Channel, m.Version)),
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
	reversed = append(reversed, m.MetricName)
	reversed = append(reversed, fmt.Sprintf("%d", m.MetricValue))
	for i := len(orig) - 1; i >= 2; i-- {
		reversed = append(reversed, orig[i])
	}
	return reversed
}

// version represents a Brave Browser version, which is based on semantic
// versioning.
type version struct {
	major int
	minor int
	patch int
}

// newerThan returns true if the given version is newer than the object's
// version.
func (v1 *version) newerThan(v2 *version) bool {
	if v1.major > v2.major {
		return true
	}
	if v1.major < v2.major {
		return false
	}
	// Major version numbers are identical.
	if v1.minor > v2.minor {
		return true
	}
	if v1.minor < v2.minor {
		return false
	}
	// Minor version numbers are identical.
	if v1.patch > v2.patch {
		return true
	}
	return false
}

// isEqual returns true if the given version is identical to the object's
// version.
func (v1 *version) isEqual(v2 *version) bool {
	return v1.major == v2.major && v1.minor == v2.minor && v1.patch == v2.patch
}

// newVersion returns a new version for the given version string.
func newVersion(strVersion string) *version {
	var err error
	attrs := strings.Split(strVersion, ".")
	v := &version{}

	v.major, err = strconv.Atoi(attrs[0])
	if err != nil {
		elog.Fatalf("Couldn't convert major version number %s to int.", attrs[0])
	}
	v.minor, err = strconv.Atoi(attrs[1])
	if err != nil {
		elog.Fatalf("Couldn't convert minor version number %s to int.", attrs[1])
	}
	v.patch, err = strconv.Atoi(attrs[2])
	if err != nil {
		elog.Fatalf("Couldn't convert patch version number %s to int.", attrs[2])
	}

	return v
}

// isRecentVersion returns true if the given version is identical to or newer
// than the latest version we've seen so far for the given channel.  Note that
// the function updates the latest versions as it's seeing newer versions.  The
// fact that we update the latest version as we're going through measurements
// means that we will have a small number of false positives but that doesn't
// matter considering that we're processing millions of measurements.
func isRecentVersion(channel, strVersion string) bool {
	maybeLastVersion, exists := lastVersion[channel]
	if !exists {
		elog.Printf("Got unexpected channel %q.", channel)
		return false
	}
	if strVersion == "" {
		return false
	}

	version := newVersion(strVersion)
	if version.newerThan(maybeLastVersion) {
		elog.Printf("Updating latest version for %s to %s.", channel, strVersion)
		lastVersion[channel] = version
		return true
	}
	return version.isEqual(maybeLastVersion)
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
