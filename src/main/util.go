package main

import (
	"time"
)

// Create a channel which will receive a value of 'true'
// every `ms` milliseconds. Not the same as sync.Ticker, which
// does not wait for a receiver - this interval channel will
// 'wait' for a receiver and then resume the interval. Sending
// a signal back the other direction (either `true` or `false`)
// will close down the timer (but NOT close the channel)
func SetInterval(ms int) chan bool {
	signal := make(chan bool)
	go func() {
		for {
			time.Sleep(1 * time.Millisecond)
			select {
			case <-signal:
				return
			case signal <- true:
				continue
			}
		}
	}()
	return signal
}
