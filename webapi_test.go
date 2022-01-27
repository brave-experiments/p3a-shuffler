package main

import (
	"bytes"
	"testing"
)

func TestShufflerMessage(t *testing.T) {
	_ = ShufflerMessage{}
}

func TestP3AMessage(t *testing.T) {
	m1 := P3AMessage{YearOfSurvey: 2022}
	m2 := P3AMessage{YearOfSurvey: 2021}
	m1CrowdID := m1.CrowdID()
	m2CrowdID := m2.CrowdID()

	if m1CrowdID == m2CrowdID {
		t.Error("Crowd ID of two distinct messages must not be identical.")
	}
	if m1.String() == m2.String() {
		t.Error("String representation of two distinct messages must not be identical.")
	}
	if bytes.Equal(m1.Payload(), m2.Payload()) {
		t.Error("Payload of two distinct messages must not be identical.")
	}

	if m1CrowdID != m1.CrowdID() {
		t.Error("Crowd ID of two identical messages must not differ.")
	}
}
