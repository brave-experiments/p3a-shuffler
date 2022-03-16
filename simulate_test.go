package main

import "testing"

func TestEntropy(t *testing.T) {
	highEntropy := map[string]int{
		"0": 1,
		"1": 1,
		"2": 1,
		"3": 1,
		"4": 1,
		"5": 1,
		"6": 1,
		"7": 1,
	}

	e := empiricalEntropy(highEntropy)
	if e != float64(1) {
		t.Fatalf("expected maximum entropy but got %.2f.", e)
	}

	e = empiricalEntropy(map[string]int{"0": 1})
	if e != float64(0) {
		t.Fatalf("expected minimum entropy but got %.2f.", e)
	}
}
