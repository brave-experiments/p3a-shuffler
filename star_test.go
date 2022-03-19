package main

import "testing"

func initSTAR() *NestedSTAR {
	star := NewNestedSTAR(&simulationConfig{})

	star.root.Add([]string{"baz"})
	star.root.Add([]string{"bar"})
	star.root.Add([]string{"foo", "bar"})
	star.root.Add([]string{"foo", "baz"})
	star.root.Add([]string{"foo", "bar", "baz"})
	star.root.Add([]string{"qux", "foo", "bar", "qux"})
	star.root.Add([]string{"qux", "foo", "bar", "bar"})

	return star
}

func isMapEqual(m1, m2 map[int]int) bool {
	if len(m1) != len(m2) {
		return false
	}

	for key, v1 := range m1 {
		v2, exists := m2[key]
		if !exists {
			return false
		}
		if v1 != v2 {
			return false
		}
	}
	return true
}

func TestNumLeafs(t *testing.T) {
	var n, expected int
	star := initSTAR()

	n = star.root.NumNodes()
	expected = 6
	if n != expected {
		t.Fatalf("expected %d but got %d nodes in tree.", expected, n)
	}

	n = star.root.NumTags()
	expected = 11
	if n != expected {
		t.Fatalf("expected %d but got %d tags in tree.", expected, n)
	}

	n = star.root.NumLeafTags()
	expected = 6
	if n != expected {
		t.Fatalf("expected %d but got %d leaf tags in tree.", expected, n)
	}
}

func TestAggregate(t *testing.T) {
	star := initSTAR()
	state := star.root.Aggregate(4, 1, []string{})

	expectedFull, expectedPartial := 2, 4
	if state.FullMsmts != expectedFull {
		t.Fatalf("expected %d but got %d full measurements.", expectedFull, state.FullMsmts)
	}
	if state.PartialMsmts != expectedPartial {
		t.Fatalf("expected %d but got %d partial measurements.", expectedPartial, state.PartialMsmts)
	}

	expectedLens := map[int]int{
		1: 2,
		2: 1,
		3: 1,
	}
	if !isMapEqual(state.LenPartialMsmts, expectedLens) {
		t.Fatalf("expected %v but got %v.", expectedLens, state.LenPartialMsmts)
	}
}

func TestAggregationState(t *testing.T) {
	s1 := NewAggregationState()
	s2 := NewAggregationState()

	if s1.AnythingUnlocked() {
		t.Fatal("expected no measurements to be unlocked for empty state")
	}

	s1.AddLenTags(0, 10)
	s1.AddLenTags(1, 5)
	s2.AddLenTags(1, 15)
	s1.Augment(s2)
	if s1.LenPartialMsmts[0] != 10 {
		t.Fatalf("expected 10 but got %d", s1.LenPartialMsmts[0])
	}
	if s1.LenPartialMsmts[1] != 20 {
		t.Fatalf("expected 20 but got %d", s1.LenPartialMsmts[0])
	}
}
