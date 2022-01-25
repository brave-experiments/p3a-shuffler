package main

import (
	"log"
	"sync"
)

type Decrypter struct {
	sync.WaitGroup
	inbox  chan []byte
	outbox chan Report
	done   chan bool
}

func (d *Decrypter) Start() {
	d.Add(1)
	go func() {
		defer d.Done()
		for {
			select {
			case <-d.done:
				return
			case r := <-d.inbox:
				// TODO
				log.Println(r)
			}
		}
	}()
}

func (d *Decrypter) Stop() {
	d.done <- true
	d.Wait()
}

func (d *Decrypter) Decrypt(payload []byte) ([]byte, error) {
	return []byte{}, nil
}

func NewDecrypter() *Decrypter {
	return &Decrypter{
		inbox:  make(chan []byte),
		outbox: make(chan Report),
	}
}
