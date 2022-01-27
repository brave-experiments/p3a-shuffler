package main

import (
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
	CrowdID() CrowdID
	Payload() []byte
}

// Shuffler implements four tasks: anonymization, shuffling, thresholding, and
// batching.
type Shuffler struct {
	sync.WaitGroup
	inbox       chan Report
	outbox      chan []Report
	done        chan bool
	BatchPeriod time.Duration
	briefcase   *Briefcase
}

// NewShuffler returns a new shuffler that batches reports until the given
// batch period.
func NewShuffler(batchPeriod time.Duration) *Shuffler {
	return &Shuffler{
		inbox:       make(chan Report),
		outbox:      make(chan []Report),
		done:        make(chan bool),
		BatchPeriod: batchPeriod,
		briefcase:   NewBriefcase(),
	}
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
				if s.briefcase.NumCrowdIDs() == 0 {
					break
				}
				s.briefcase.DumpFewerThan(10)
				reports, err := s.briefcase.ShuffleAndEmpty()
				if err != nil {
					log.Printf("Shuffler: Briefcase failed to shuffle: %s", err)
					break
				}
				s.outbox <- reports
				log.Printf("Shuffler: Sent %d reports to outbox.", len(reports))
			}
		}
	}()
}

// Stop stops the shuffler.
func (s *Shuffler) Stop() {
	s.done <- true
	s.Wait()
}
