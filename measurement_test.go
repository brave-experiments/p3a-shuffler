package main

import (
	"strings"
	"testing"
)

var m P3AMeasurement = P3AMeasurement{
	YearOfSurvey:  2022,
	YearOfInstall: 2022,
	WeekOfSurvey:  1,
	WeekOfInstall: 1,
	MetricValue:   1,
	MetricName:    "Brave.Rewards.WalletState",
	CountryCode:   "CA",
	Platform:      "winx64-bc",
	Version:       "1.37.60",
	Channel:       "nightly",
	RefCode:       "none",
}

func TestCrowdIDs(t *testing.T) {
	fullCrowdID1 := m.CrowdID(attrsAll)
	originCrowdID1 := m.CrowdID(attrsNoValue)

	if fullCrowdID1 == originCrowdID1 {
		t.Fatalf("Full and origin crowd ID are unlikely to be identical.")
	}

	m.MetricValue++
	fullCrowdID2 := m.CrowdID(attrsAll)
	originCrowdID2 := m.CrowdID(attrsNoValue)
	if originCrowdID1 != originCrowdID2 {
		t.Fatalf("Origin crowd ID must not be affected when metric value changes.")
	}
	if fullCrowdID2 == fullCrowdID1 {
		t.Fatalf("Full crowd ID must be affected when metric value changes.")
	}
}

func TestOrdering(t *testing.T) {
	hef := m.OrderHighEntropyFirst(attrsAll)
	hel := m.OrderHighEntropyLast(attrsAll)

	if len(hef) != 11 || len(hel) != 11 {
		t.Fatalf("Measurement doesn't have expected number of attributes.")
	}

	if len(hef) != len(hel) {
		t.Fatalf("Ordered measurements don't have the same length.")
	}

	for i := range hef {
		if hef[i] != hel[len(hel)-1-i] {
			t.Fatalf("HEF and HEL ordering don't match.")
		}
	}
}

func TestIsValid(t *testing.T) {
	if !m.IsValid() {
		t.Fatal("Standard measurement considered not valid.")
	}
	badM := m
	badM.MetricName = ""
	if badM.IsValid() {
		t.Fatal("Bad measurement considered valid despite not having metric name.")
	}
}

func TestCSV(t *testing.T) {
	header := m.CSVHeader()
	record := m.CSV()
	if strings.Count(header, ",") != strings.Count(record, ",") {
		t.Fatal("CSV header and record don't have the same number of commas.")
	}
}
