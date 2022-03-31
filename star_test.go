package main

import "testing"

func initFakeSTAR() *NestedSTAR {
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

func initSTAR() (*NestedSTAR, int, int) {
	star := NewNestedSTAR(&simulationConfig{})

	maxTags, threshold := 3, 5
	// A few full measurements that meet our k=5 threshold.
	star.root.Add([]string{"US", "release", "windows"})
	star.root.Add([]string{"US", "release", "windows"})
	star.root.Add([]string{"US", "release", "windows"})
	star.root.Add([]string{"US", "release", "windows"})
	star.root.Add([]string{"US", "release", "windows"})
	// Two partial measurements (consisting of ["US", "release"] that don't
	// meet k=5.
	star.root.Add([]string{"US", "release", "linux"})
	star.root.Add([]string{"US", "release", "linux"})
	star.root.Add([]string{"US", "release", "macos"})
	// Two partial measurements (consisting of ["US"]).
	star.root.Add([]string{"US", "nightly", "windows"})
	star.root.Add([]string{"US", "beta", "windows"})
	// One lost measurement.
	star.root.Add([]string{"CA", "release", "windows"})

	return star, maxTags, threshold
}

func TestRealSTAR(t *testing.T) {
	star, maxTags, threshold := initSTAR()
	state := star.root.Aggregate(maxTags, threshold, []string{})

	expectedFull, expectedPartial := 5, 5
	if state.FullMsmts != expectedFull {
		t.Fatalf("expected %d but got %d full measurements.", expectedFull, state.FullMsmts)
	}
	if state.PartialMsmts != expectedPartial {
		t.Fatalf("expected %d but got %d partial measurements.", expectedPartial, state.PartialMsmts)
	}

	expectedLens := map[int]int{
		1: 2,
		2: 3,
	}
	if !isMapEqual(state.LenPartialMsmts, expectedLens) {
		t.Fatalf("expected %v but got %v.", expectedLens, state.LenPartialMsmts)
	}
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
	star, _, _ := initSTAR()

	n = star.root.NumNodes()
	expected = 7
	if n != expected {
		t.Fatalf("expected %d but got %d nodes in tree.", expected, n)
	}

	n = star.root.NumTags()
	expected = 12
	if n != expected {
		t.Fatalf("expected %d but got %d tags in tree.", expected, n)
	}

	n = star.root.NumLeafTags()
	expected = 6
	if n != expected {
		t.Fatalf("expected %d but got %d leaf tags in tree.", expected, n)
	}
}

func TestAggregationState(t *testing.T) {
	s1 := NewAggregationState()
	s2 := NewAggregationState()

	if !s1.NothingUnlocked() {
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
