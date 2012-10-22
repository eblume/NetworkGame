package main

import (
	"time"
)

// The core package 'time' actually defines what they call a 'ticker' that is
// very similar to this, but different in subtle ways. In particular, a ticker
// will adjust it's interval to accomodate slow readers. This is not what we
// want at ALL. (As a small point of pride, it is also needlessly complicated.)

// Define an Interval channel that will get a boolean 'pulse' every 'duration'
// milliseconds until someone sends a message from the other side to 'close'.
func Interval(duration int, close chan bool) chan bool {
	// TODO - got to be a better way to close this.
	pulse := make(chan bool)
	go func() {
		for {
			time.Sleep(time.Duration(duration) * time.Millisecond)
			select {
			case <-close:
				return
			case pulse <- true:
				continue
			}
		}
	}()
	return pulse
}
