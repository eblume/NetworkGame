package main

import (
	"sync"
	"time"
)

// The type for 'prototyped' (typeless) variables. This is useful
// for when we can programmatically infer types. You know, dynamic
// typing. The thing *real* programmers use and love. ;)
type Proto interface{}

// Gather all signals on `recv`, and pack them in to a slice.
// Think of this as 'Channel->Slice'. Type assertions will be
// needed on the values that this returns. The `recv` channel
// MUST be close()'ed or else this function will block forever.
func Gather(recv chan Proto) []Proto {
	result := make([]Proto, 0, 1)
	for val := range recv {
		result = append(result, val)
	}
	return result
}

// Gather up to `count` items at a time from `recv`, and send them
// in one group. If the channel closes before `count` items are
// gathered, just send that (shortened) group. Never sends an
// empty group - instead, closes the channel when done.
// This happens without blocking, unlike `Gather`.
func GatherN(recv chan Proto, count int) chan []Proto {
	send := make(chan []Proto)
	go func() {
		tuple := make([]Proto, 0, count)
		for val := range recv {
			tuple = append(tuple, val)
			if len(tuple) >= count {
				send <- tuple
				tuple = make([]Proto, 0, count)
			}
		}
		// If a partial tuple exists, send it too
		if len(tuple) > 0 {
			send <- tuple
		}
		close(send)
	}()
	return send
}

// Inverse of 'Gather' - put all the values in `vals` on a new
// channel. Think of it as "Slice->Channel". Unlike Gather, does
// not block. If `vals` is empty, the returned channel behaves
// as expected - that is, it will simply be closed (or close
// very quickly.) Order is preserved.
func Send(vals []Proto) chan Proto {
	send := make(chan Proto, len(vals))
	go func() {
		for i := range vals {
			send <- vals[i]
		}
		close(send)
	}()
	return send
}

// Send all the elements of a to b, asynchronously, and then
// close b.
func Splice(a chan Proto, b chan Proto) {
	go func() {
		for val := range a {
			b <- val
		}
		close(b)
	}()
}

// Reducing function type definition - implementer will need
// to dilligently box & unbox from Proto via Type Assertion
type ReduceFn func(Proto, Proto) Proto

// Reduce the `recv` channel by applying `fn` on each
// value in the channel subsequently, with the first argument to
// `fn` being either the first received element or the previous
// result of each thereafter application of `fn`. The first
// received element does not have `fn` called on it. Note that
// though the returned object is a channel, only one value
// will ever be sent on it (the result value).
func Reduce(fn ReduceFn, recv chan Proto) chan Proto {
	var accum Proto
	send := make(chan Proto, 1)
	accum = nil
	go func() {
		for val := range recv {
			if accum == nil {
				accum = val
			} else {
				accum = fn(accum, val)
			}
		}
		send <- accum
		close(send)
	}()
	return send
}

// Filter function type dfinition - implemented will need
type FilterFn func(Proto) bool

// Filter the channel with the given function. The function
// must return true or false for each individual element the
// channel may receive. If true, the element will be sent on
// the return channel.
func Filter(fn FilterFn, recv chan Proto) chan Proto {
	send := make(chan Proto)
	go func() {
		for val := range recv {
			if fn(val) {
				send <- val
			}
		}
		close(send)
	}()
	return send
}

// Combine multiple input channels in to one.
func Multiplex(inputs []chan Proto) chan Proto {
	send := make(chan Proto)
	go func() {
		var group sync.WaitGroup
		for i := range inputs {
			group.Add(1)
			go func(input chan Proto) {
				for val := range input {
					send <- val
				}
				group.Done()
			}(inputs[i])
		}
		go func() {
			group.Wait()
			close(send)
		}()
	}()
	return send
}

// Seperate an input channel in to two output channels,
// By applying the filter function. The first output channel
// will get the values that passed the filter, the second will
// get those that did not.
func Demultiplex(fn FilterFn, recv chan Proto) []chan Proto {
	send := make([]chan Proto, 2)
	send[0] = make(chan Proto)
	send[1] = make(chan Proto)
	go func() {
		for val := range recv {
			if fn(val) {
				send[0] <- val
			} else {
				send[1] <- val
			}
		}
		close(send[0])
		close(send[1])
	}()
	return send
}

// Mapping function type definition.
type MapFn func(Proto) Proto

// Apply `fn` to each value on `recv`, and send the results on.
// Order is preserved. Though `Map` does not block, it is not
// parallel - for a parallel version, see `PMap`.
func Map(fn MapFn, recv chan Proto) chan Proto {
	send := make(chan Proto)
	go func() {
		for val := range recv {
			send <- fn(val)
		}
		close(send)
	}()
	return send
}

// Exactly like `Map`, but every application gets its own
// goroutine. In general, only use this if the mapped function
// is 'heavy' enough to deserve parallelization, else see `Map`.
// Order will NOT be preserved.
func PMap(fn MapFn, recv chan Proto) chan Proto {
	send := make(chan Proto)
	internal := make(chan Proto)
	go func() {
		seen := 0
		for val := range recv {
			seen++
			go func(value Proto) {
				internal <- fn(value)
			}(val)
		}
		for i := 0; i < seen; i++ {
			send <- <-internal
		}
		close(send)
	}()
	return send
}

// Helper struct for `PFilter`
type filterResult struct {
	passed bool
	val    Proto
}

// Exactly like `Filter`, but every filter application gets its
// own goroutine. See `PMap` for details. Order NOT preserved.
func PFilter(fn FilterFn, recv chan Proto) chan Proto {
	send := make(chan Proto)
	internal := make(chan filterResult)
	go func() {
		seen := 0
		for val := range recv {
			if val == nil {
				continue
			}
			seen++
			go func(value Proto) {
				internal <- filterResult{
					passed: fn(value),
					val:    value,
				}
			}(val)
		}
		for i := 0; i < seen; i++ {
			val := <-internal
			if val.passed {
				send <- val.val
			}
		}
		close(send)
	}()
	return send
}

// Helper functiopn for PReduce.
// Like 'Multiplex', except that it only reads a certain
// number of elements from b (reliant on the number of elements
// from a) and then stops, not waiting for b to close.
func combine(a chan Proto, b chan Proto) chan Proto {
	send := make(chan Proto)
	go func() {
		// First send all from a, while counting
		count := 0
		for val := range a {
			count++
			send <- val
		}
		// Next, send count-1 times from b
		for i := 0; i < count-1; i++ {
			send <- <-b
		}
		close(send)
	}()
	return send
}

// Helper function for PReduce.
// Receives a single value from a, sends it to b, then closes b.
func output(a chan Proto, b chan Proto) {
	go func() {
		partial := (<-a).([]Proto)
		b <- partial[0]
		close(b)
	}()
}

// Helper function for PReduce
// Return 'true' if this is a 'full group', meaning it has 2 elems
func filtPartial(a Proto) bool {
	pair := a.([]Proto)
	return len(pair) == 2
}

// Exactly like `Reduce`, but every reduction is performed in
// its own goroutine. `fn` MUST BE FULLY ASSOSCIATIVE - that is,
// the order of its arguments absolutely can't matter. If 
// `fn` can't be made assosciative, you MUST use `Reduce` instead!
func PReduce(fn ReduceFn, recv chan Proto) chan Proto {
	send := make(chan Proto)
	splice := make(chan Proto)
	combined := combine(recv, splice)
	gathered := GatherN(combined, 2)
	pairs, partial := Demultiplex(filtPartial, gathered)
	output(partial, send)
	reducer := func(a Proto) Proto {
		pair := a.([]Proto)
		return fn(a[0], a[1])
	}
	reduced := PMap(reducer, pairs)
	Splice(reduced, splice)
	return send
}

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
