package main

import (
	"sync"
	"time"
)

type empty struct{}

type CrowdID string

type Report struct {
	CrowdID CrowdID
	Payload []byte
}

// Shuffler implements four tasks: anonymization, shuffling, thresholding, and
// batching.
type Shuffler struct {
	sync.WaitGroup
	inbox       chan *Report
	outbox      chan []*Report
	done        chan bool
	BatchPeriod time.Duration
	briefcase   *Briefcase
}

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
				s.briefcase.DumpFewerThan(10)
				reports, _ := s.briefcase.Shuffle()
				s.outbox <- reports
			}
		}
	}()
}

func (s *Shuffler) Stop() {
	s.done <- true
	s.Wait()
}

func NewShuffler(batchPeriod time.Duration) *Shuffler {
	return &Shuffler{
		inbox:       make(chan *Report),
		outbox:      make(chan []*Report),
		done:        make(chan bool),
		BatchPeriod: batchPeriod,
		briefcase:   NewBriefcase(),
	}
}
