package main

import (
	"fmt"
	"testing"
)

type DummyReport struct {
	crowdID CrowdID
	payload []byte
}

func (d DummyReport) CrowdID() CrowdID {
	return d.crowdID
}
func (d DummyReport) Payload() []byte {
	return d.payload
}

func getFullBriefcase(reports, crowdIDs int) *Briefcase {
	b := NewBriefcase()

	for i := 0; i < reports; i++ {
		b.Add(&DummyReport{
			crowdID: CrowdID(fmt.Sprintf("%d", i%crowdIDs)),
			payload: []byte("foo"),
		})
	}
	return b
}

func checkLengths(t *testing.T, b *Briefcase, numReports, numCrowdIDs int) {
	if b.NumReports() != numReports {
		t.Fatalf("Expected %d reports but got %d.", numReports, b.NumReports())
	}
	if b.NumCrowdIDs() != numCrowdIDs {
		t.Fatalf("Expected %d crowd IDs but got %d.", numCrowdIDs, b.NumCrowdIDs())
	}
}

func TestAdd(t *testing.T) {
	numReports, numCrowdIDs := 100, 2
	b := getFullBriefcase(numReports, numCrowdIDs)
	checkLengths(t, b, numReports, numCrowdIDs)

	// Now empty our suitcase.
	b.Empty()
	numReports, numCrowdIDs = 0, 0
	checkLengths(t, b, numReports, numCrowdIDs)
}

func TestDumpFewerThan(t *testing.T) {
	numReports, numCrowdIDs := 100, 2
	b := getFullBriefcase(numReports, numCrowdIDs)

	// Add two reports that are part of the same crowd ID.
	b.Add(&DummyReport{crowdID: CrowdID("foo"), payload: []byte("bar")})
	b.Add(&DummyReport{crowdID: CrowdID("foo"), payload: []byte("bar")})
	numReports += 2
	numCrowdIDs++

	// Nothing should change here.
	b.DumpFewerThan(2)
	checkLengths(t, b, numReports, numCrowdIDs)

	// Our two latest reports should now be deleted.
	b.DumpFewerThan(3)
	numReports -= 2
	numCrowdIDs--
	checkLengths(t, b, numReports, numCrowdIDs)

	// All remaining reports should now be deleted.
	b.DumpFewerThan(100)
	numReports, numCrowdIDs = 0, 0
	checkLengths(t, b, numReports, numCrowdIDs)
}

func TestShuffleAndEmpty(t *testing.T) {
	numReports, numCrowdIDs := 100, 2
	b := getFullBriefcase(numReports, numCrowdIDs)

	reports1, err := b.ShuffleAndEmpty()
	if err != nil {
		t.Fatalf("Failed to shuffle reports: %v", err)
	}
	if b.NumCrowdIDs() != 0 {
		t.Fatalf("Failed to empty the briefcase after shuffling.")
	}

	b = getFullBriefcase(numReports, numCrowdIDs)
	reports2, err := b.ShuffleAndEmpty()
	if err != nil {
		t.Fatalf("Failed to shuffle reports: %v", err)
	}

	// Randomness is notoriously hard to unit test.  We therefore check if the
	// two shuffled reports are identical.  Admittedly, that's a low bar to
	// clear but better than nothing.
	for i := range reports1 {
		if reports1[i] != reports2[i] {
			break
		} else if i == len(reports1) {
			t.Fatalf("Two shuffled reports are identical.")
		}
	}
}
