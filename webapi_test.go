package main

import (
	"bytes"
	"testing"
)

func TestShufflerMeasurement(t *testing.T) {
	_ = ShufflerMeasurement{}
}

func TestP3AMeasurement(t *testing.T) {
	m1 := P3AMeasurement{YearOfSurvey: 2022}
	m2 := P3AMeasurement{YearOfSurvey: 2021}
	m1CrowdID := m1.CrowdID(defaultCrowdIDMethod)
	m2CrowdID := m2.CrowdID(defaultCrowdIDMethod)

	if m1CrowdID == m2CrowdID {
		t.Error("Crowd ID of two distinct measurements must not be identical.")
	}
	if m1.String() == m2.String() {
		t.Error("String representation of two distinct measurements must not be identical.")
	}
	if bytes.Equal(m1.Payload(), m2.Payload()) {
		t.Error("Payload of two distinct measurements must not be identical.")
	}

	if m1CrowdID != m1.CrowdID(defaultCrowdIDMethod) {
		t.Error("Crowd ID of two identical measurements must not differ.")
	}
}
