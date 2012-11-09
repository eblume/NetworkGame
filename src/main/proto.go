package main

import (
    "sync"
)

// The type for 'prototyped' (typeless) variables. This is useful
// for when we can programmaticly infer types. You know, dynamic
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

// Inverse of 'Gather' - put all the values in `vals` on a new
// channel. Think of it as "Slice->Channel". Unlike Gather, does
// not block. If `vals` is empty, the returned channel behaves
// as expected - that is, it will simply be closed (or close
// very quickly.) Order is preserved.
func Send(vals []Proto) chan Proto {
    send := make(chan Proto, len(vals))
    go func() {
        defer close(send)
        for i := range vals {
            send <- vals[i]
        }
    }()
    return send
}

// Send all the elements of a to b, asynchronously, and then
// close b.
func Splice(a chan Proto, b chan Proto) {
    go func() {
        defer close(b)
        for val := range a {
            b <- val
        }
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
    send := make(chan Proto, 1)
    go func() {
        defer close(send)
        var accum Proto = nil
        for val := range recv {
            if accum == nil {
                accum = val
            } else {
                accum = fn(accum, val)
            }
        }
        send <- accum
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
        defer close(send)
        for val := range recv {
            if fn(val) {
                send <- val
            }
        }
    }()
    return send
}

// Combine multiple input channels in to one.
func Multiplex(inputs ...chan Proto) chan Proto {
    send := make(chan Proto)
    go func() {
        defer close(send)
        var group sync.WaitGroup
        defer group.Wait()
        for i := range inputs {
            group.Add(1)
            go func(input chan Proto) {
                defer group.Done()
                for val := range input {
                    send <- val
                }
            }(inputs[i])
        }
    }()
    return send
}

// Seperate an input channel in to two output channels,
// By applying the filter function. The first output channel
// will get the values that passed the filter, the second will
// get those that did not.
func Demultiplex(fn FilterFn, recv chan Proto) (chan Proto, chan Proto) {
    passed := make(chan Proto)
    failed := make(chan Proto)
    go func() {
        defer close(passed)
        defer close(failed)
        for val := range recv {
            if fn(val) {
                passed <- val
            } else {
                failed <- val
            }
        }
    }()
    return passed, failed
}

// Mapping function type definition.
type MapFn func(Proto) Proto

// Apply `fn` to each value on `recv`, and send the results on.
// Order is preserved. Though `Map` does not block, it is not
// parallel - for a parallel version, see `PMap`.
func Map(fn MapFn, recv chan Proto) chan Proto {
    send := make(chan Proto)
    go func() {
        defer close(send)
        for val := range recv {
            send <- fn(val)
        }
    }()
    return send
}

// Exactly like `Map`, but every application gets its own
// goroutine. In general, only use this if the mapped function
// is 'heavy' enough to deserve parallelization, else see `Map`.
// Order will NOT be preserved.
func PMap(fn MapFn, recv chan Proto) chan Proto {
    send := make(chan Proto)
    go func() {
        defer close(send)
        var group sync.WaitGroup
        defer group.Wait()
        for val := range recv {
            group.Add(1)
            go func(value Proto) {
                defer group.Done()
                send <- fn(value)
            }(val)
        }
    }()
    return send
}

// Exactly like `Filter`, but every filter application gets its
// own goroutine. See `PMap` for details. Order NOT preserved.
func PFilter(fn FilterFn, recv chan Proto) chan Proto {
    send := make(chan Proto)
    go func() {
        defer close(send)
        var group sync.WaitGroup
        defer group.Wait()
        for val := range recv {
            group.Add(1)
            go func(value Proto) {
                defer group.Done()
                if fn(value) {
                    send <- value
                }
            }(val)
        }
    }()
    return send
}

// Why no PReduce? I spent some time working on it but couldn't ever seem to
// get it to work right. I had tried the approach of a circular link around
// PMap, where input to PMap was a pair of elements that got 'reduced' to
// one output element that then went back around again until only one remained.
// In theory this should work, although there may be intrinsic limitations to
// the parallelism in that case.
