package main

import "sync"

type multiplexer struct {
	sync.Mutex
	origChan chan []Report
	fwdChans []chan []Report
}

func newMultiplexer() *multiplexer {
	return &multiplexer{
		origChan: make(chan []Report),
	}
}

func (m *multiplexer) register(c chan []Report) {
	m.Lock()
	defer m.Unlock()
	elog.Println("Registering channel.")
	m.fwdChans = append(m.fwdChans, c)
}

func (m *multiplexer) start() {
	elog.Printf("Starting to multiplex to %d channels.", len(m.fwdChans))
	go func() {
		for rs := range m.origChan {
			m.Lock()
			for _, c := range m.fwdChans {
				c <- rs
			}
			m.Unlock()
		}
		elog.Println("Original channel closed.")
		m.stop()
	}()
}

func (m *multiplexer) stop() {
	m.Lock()
	defer m.Unlock()
	elog.Println("Stopping multiplexing.")
	for _, c := range m.fwdChans {
		close(c)
	}
}
