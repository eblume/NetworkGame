package main

import (
	"testing"
	"time"
)

func TestInterval(t *testing.T) {
	ival := SetInterval(1)

	for count := 0; count < 3; count++ {
		<-ival
	}
	ival <- true
	// Wait a bit - if closing 'fails' then there will be a message
	// waiting for us.
	time.Sleep(3 * time.Millisecond)
	select {
	case <-ival:
		t.Error("Failed to close interval.")
	default:
	}
}
