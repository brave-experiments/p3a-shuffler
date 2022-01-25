package main

import (
	"crypto/rand"
	"math/big"
	"sync"
)

// Briefcase contains reports.  Obviously!
type Briefcase struct {
	sync.Mutex
	Reports map[CrowdID][]*Report
}

func NewBriefcase() *Briefcase {
	return &Briefcase{
		Reports: make(map[CrowdID][]*Report),
	}
}

func (b *Briefcase) Empty() {
	b.Lock()
	defer b.Unlock()

	b.Reports = make(map[CrowdID][]*Report)
}

func (b *Briefcase) NumCrowdIDs() int {
	b.Lock()
	defer b.Unlock()

	return len(b.Reports)
}

func (b *Briefcase) NumReports() int {
	b.Lock()
	defer b.Unlock()

	num := 0
	for _, reports := range b.Reports {
		num += len(reports)
	}
	return num
}

// Shuffle gives the briefcase a good shuffle.
func (b *Briefcase) Shuffle() ([]*Report, error) {
	b.Lock()
	defer b.Unlock()

	result := []*Report{}
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

	return result, nil
}

// DumpFewerThan dumps all reports (as identified by CrowdID) fewer than the
// given minimum amount, e.g., if min equals 5, we remove all reports whose
// total number of CrowdID is fewer than 5.
func (b *Briefcase) DumpFewerThan(min int) {
	b.Lock()
	defer b.Unlock()

	for crowdID, reports := range b.Reports {
		// We don't have the minimum number of reports for the given crowd ID.
		// Discard all the reports.
		if len(reports) < min {
			delete(b.Reports, crowdID)
		}
	}
}

// Add adds a new report to the briefcase.
func (b *Briefcase) Add(r *Report) {
	b.Lock()
	defer b.Unlock()

	reports, exists := b.Reports[r.CrowdID]
	if !exists {
		b.Reports[r.CrowdID] = []*Report{r}
	} else {
		b.Reports[r.CrowdID] = append(reports, r)
	}
}
