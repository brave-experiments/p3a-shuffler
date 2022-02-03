package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// CrowdID represents the crowd ID of a given report.
type CrowdID string

// Report defines an interface that represents a report in our briefcase.  A
// report must be able to return its crowd ID and payload; and it must be
// marshal-able.
type Report interface {
	CrowdID(method int) CrowdID
	Payload() []byte
}

// Shuffler implements four tasks: anonymization, shuffling, thresholding, and
// batching.
type Shuffler struct {
	sync.WaitGroup
	inbox              chan []Report
	outbox             chan []Report
	done               chan bool
	anonymityThreshold int
	BatchPeriod        time.Duration
	briefcase          *Briefcase
}

// NewShuffler returns a new shuffler that batches reports until the given
// batch period.
func NewShuffler(batchPeriod time.Duration, anonymityThreshold int, crowdIDMethod int) *Shuffler {
	return &Shuffler{
		inbox:              make(chan []Report),
		outbox:             make(chan []Report),
		done:               make(chan bool),
		anonymityThreshold: anonymityThreshold,
		BatchPeriod:        batchPeriod,
		briefcase:          NewBriefcase(crowdIDMethod),
	}
}

// String returns a summary of the shuffler's internal state.
func (s *Shuffler) String() string {
	return fmt.Sprintf("Briefcase contains %d crowd IDs; %d reports.",
		s.briefcase.NumCrowdIDs(),
		s.briefcase.NumReports())
}

// Start starts the shuffler.
func (s *Shuffler) Start() {
	s.Add(1)
	go func() {
		defer s.Done()
		ticker := time.NewTicker(s.BatchPeriod)
		for {
			select {
			case <-s.done:
				s.briefcase.Empty()
				return
			case r := <-s.inbox:
				s.briefcase.Add(r)
			case <-ticker.C:
				if err := s.endBatchPeriod(); err != nil {
					log.Printf("Shuffler: failed to end batch period because: %s", err)
				}
			}
		}
	}()
}

// endBatchPeriod does the housekeeping that's necessary once our batch period
// ends, i.e. it enforces our k-anonymity guarantees on all reports, shuffles
// the remaining reports, and empties our briefcase.  Whatever reports are left
// are then sent to the shuffler's outbox.
func (s *Shuffler) endBatchPeriod() error {
	if s.briefcase.NumCrowdIDs() == 0 {
		return nil
	}
	s.briefcase.DumpFewerThan(s.anonymityThreshold)

	reports, err := s.briefcase.ShuffleAndEmpty()
	if err != nil {
		return err
	}
	s.outbox <- reports
	log.Printf("Shuffler: Sent %d reports to outbox.", len(reports))
	return nil
}

// Stop stops the shuffler.
func (s *Shuffler) Stop() {
	s.done <- true
	s.Wait()
}
