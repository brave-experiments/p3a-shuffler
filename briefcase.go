package main

import (
	"crypto/rand"
	"math/big"
	"sync"
)

// Briefcase contains reports.  Obviously!
type Briefcase struct {
	sync.Mutex
	crowdIDMethod int
	Reports       map[CrowdID][]Report
}

// NewBriefcase creates and returns a new briefcase.
func NewBriefcase(crowdIDMethod int) *Briefcase {
	return &Briefcase{
		Reports:       make(map[CrowdID][]Report),
		crowdIDMethod: crowdIDMethod,
	}
}

// Empty empties the briefcase.
func (b *Briefcase) Empty() {
	b.Lock()
	defer b.Unlock()

	b.Reports = make(map[CrowdID][]Report)
}

// NumCrowdIDs returns the number of crowd IDs that the briefcase currently
// contains.
func (b *Briefcase) NumCrowdIDs() int {
	b.Lock()
	defer b.Unlock()

	return len(b.Reports)
}

// NumReports returns the number of reports that the briefcase currently
// contains.  For every crowd ID, there is at least one report.
func (b *Briefcase) NumReports() int {
	b.Lock()
	defer b.Unlock()

	num := 0
	for _, reports := range b.Reports {
		num += len(reports)
	}
	return num
}

// ShuffleAndEmpty gives the briefcase a good shuffle and subsequently empties it.
func (b *Briefcase) ShuffleAndEmpty() ([]Report, error) {
	b.Lock()
	defer b.Unlock()

	result := []Report{}
	for _, reports := range b.Reports {
		result = append(result, reports...)
	}

	for i := len(result) - 1; i > 0; i-- {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return nil, err
		}

		j := int(index.Int64())
		result[i], result[j] = result[j], result[i]
	}
	elog.Printf("Shuffled briefcase containing %d crowd IDs.", len(b.Reports))
	b.Reports = make(map[CrowdID][]Report)

	return result, nil
}

// DumpFewerThan dumps all reports (as identified by CrowdID) fewer than the
// given minimum amount, e.g., if min equals 5, we remove all reports whose
// total number of CrowdID is fewer than 5.
func (b *Briefcase) DumpFewerThan(min int) {
	b.Lock()
	defer b.Unlock()

	numDumped := 0
	for crowdID, reports := range b.Reports {
		// We don't have the minimum number of reports for the given crowd ID.
		// Discard all the reports.
		if len(reports) < min {
			delete(b.Reports, crowdID)
			numDumped++
		}
	}
	elog.Printf("Dumped %d crowd IDs for which we had fewer than %d reports.", numDumped, min)
}

// Add adds new reports to the briefcase.
func (b *Briefcase) Add(rs []Report) {
	b.Lock()
	defer b.Unlock()

	for _, r := range rs {
		reports, exists := b.Reports[r.CrowdID(b.crowdIDMethod)]
		if !exists {
			b.Reports[r.CrowdID(b.crowdIDMethod)] = []Report{r}
		} else {
			b.Reports[r.CrowdID(b.crowdIDMethod)] = append(reports, r)
		}
	}
}
