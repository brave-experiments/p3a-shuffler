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
	originCrowdID1 := m.CrowdID(attrsMinimal)

	if fullCrowdID1 == originCrowdID1 {
		t.Fatalf("Full and origin crowd ID are unlikely to be identical.")
	}

	m.MetricValue++
	fullCrowdID2 := m.CrowdID(attrsAll)
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
		if i < 2 {
			continue
		}
		j := len(hel) + 1 - i
		if hef[i] != hel[j] {
			t.Fatalf("HEF and HEL ordering don't match at i=%d, j=%d.", i, j)
		}
	}

	if hef[0] != hel[0] {
		t.Fatal("HEF and HEL don't have a shared metric name.")
	}
	if hef[1] != hel[1] {
		t.Fatal("HEF and HEL don't have a shared metric value.")
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

func TestVersions(t *testing.T) {
	if !newVersion("0.0.1").newerThan(newVersion("0.0.0")) {
		t.Fatal("expected 0.0.1 to be newer than 0.0.0")
	}
	if !newVersion("1.2.3").newerThan(newVersion("1.2.2")) {
		t.Fatal("expected 1.2.3 to be newer than 1.2.2")
	}
	if !newVersion("1.0.0").newerThan(newVersion("0.1.1")) {
		t.Fatal("expected 1.0.0 to be newer than 0.1.1")
	}

	if newVersion("1.0.0").newerThan(newVersion("1.0.0")) {
		t.Fatal("identical version cannot be newer")
	}
	if !newVersion("1.0.0").isEqual(newVersion("1.0.0")) {
		t.Fatal("identical version considered not identical")
	}
	if newVersion("1.0.0").isEqual(newVersion("2.0.0")) {
		t.Fatal("different version considered identical")
	}

	if !isRecentVersion("release", "0.0.1") {
		t.Fatal("first seen version not considered recent")
	}
	if !isRecentVersion("release", "0.0.2") {
		t.Fatal("newer version not considered recent")
	}
	if !isRecentVersion("release", "1.0.0") {
		t.Fatal("newer version not considered recent")
	}
	if isRecentVersion("release", "0.9.0") {
		t.Fatal("old version considered recent")
	}
	if !isRecentVersion("release", "1.0.0") {
		t.Fatal("new version not considered recent")
	}
}
