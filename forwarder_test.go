package main

import "testing"

func TestLifecycle(t *testing.T) {
	c := make(chan []Report)
	f := NewForwarder(c, "foo")
	f.Start()
	f.Stop()
}
