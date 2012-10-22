package main

import (
	"testing"
)

// "Actor" is a sample object following the Actor pattern described by
// util/call. You can use everything up to "ACTOR PATTERN TESTS" as an
// example of what this Actor pattern should look like.

// The actor pattern consists of five components, below:

// COMPONENT 1: State struct.
// The State struct contains one channel per method, and also defines
// the private (internal) state, with no semantic distinction. It's
// the responsibility of the coder not to touch the internal state
// (or to touch it in a thread-safe manner).
type Actor struct {
	method_noarg_noret chan *Call
	method_arg_noret   chan *Call
	method_noarg_ret   chan *Call
	method_arg_ret     chan *Call
	method_destruct    chan *Call
	//
	some_state bool
}

// COMPONENT 2: Constructor
// The constructor initializes a State Struct and starts up the
// dispatch routine.
func NewActor() *Actor {
	a := &Actor{
		method_noarg_noret: make(chan *Call),
		method_arg_noret:   make(chan *Call),
		method_noarg_ret:   make(chan *Call),
		method_arg_ret:     make(chan *Call),
		method_destruct:    make(chan *Call),
		//
		some_state: false,
	}
	go runActor(a)
	return a
}

// COMPONENT 3: Public Methods
// Exported thread-safe methods which transform a common call pattern
// in to the form needed for the channel methods, and possibly handling
// transformation of the returned values and/or synchronization.
//
// These are optional - callers can do their own Call boxing/unboxing
// if they want, which is often useful for very simple arguments/returns.

func (a *Actor) NoargNoret() {
	a.method_noarg_noret <- &Call{}
}

func (a *Actor) ArgNoret(arg1 int, arg2 int) {
	a.method_arg_noret <- &Call{
		Args: []interface{}{arg1, arg2},
	}
}

func (a *Actor) NoargRet(done chan bool) chan bool {
	// First create an interface{} receiver that we will later unbox
	go func() { // Send the method in a goroutine to avoid blocking caller
		recv := make(chan interface{})
		a.method_noarg_ret <- &Call{
			Done: recv,
		}
		done <- (<-recv).(bool) // Unbox the result.
	}()
	return done // Convenience return of the finish stream
}

func (a *Actor) ArgRet(arg1 int, arg2 int, done chan int) chan int {
	go func() {
		recv := make(chan interface{})
		a.method_arg_ret <- &Call{
			Args: []interface{}{arg1, arg2},
			Done: recv,
		}
		done <- (<-recv).(int)
	}()
	return done
}

func (a *Actor) Destruct(done chan bool) chan bool {
	go func() {
		recv := make(chan interface{})
		a.method_destruct <- &Call{
			Done: recv,
		}
		done <- (<-recv).(bool)
	}()
	return done
}

// COMPONENT 4: Dispatch Routine
// Started by the constructor, this is a goroutine that listens
// for the method channels for calls, and dispatches those calls.
func runActor(a *Actor) {
	for {
		select {
		case call := <-a.method_noarg_noret:
			methodNoargNoret(a, call)
		case call := <-a.method_arg_noret:
			methodArgNoret(a, call)
		case call := <-a.method_noarg_ret:
			methodNoargRet(a, call)
		case call := <-a.method_arg_ret:
			methodArgRet(a, call)
		case call := <-a.method_destruct:
			destructActor(a, call)
			return
		}
	}
}

// COMPONENT 5: Coroutines
// The routines that the Dispatch routine will call to handle method Call's

func methodNoargNoret(a *Actor, call *Call) {
	// Do Work
	a.some_state = true
}

func methodArgNoret(a *Actor, call *Call) {
	// The programmer knows there are two integer arguments to unpack
	x := call.Args[0].(int)
	y := call.Args[1].(int)
	// Do Something with x and y
	x = y
	y = x
}

func methodNoargRet(a *Actor, call *Call) {
	// Always wrap returns in goroutines to avoid blocking
	some_state_at_call_time := a.some_state
	go func() {
		// In general, DO NOT access state from outside of the actor's own
		// goroutine. Instead, do something like this, where we save and
		// close the state first.
		call.Done <- some_state_at_call_time
	}()
}

func methodArgRet(a *Actor, call *Call) {
	x := call.Args[0].(int)
	y := call.Args[1].(int)
	go func() {
		call.Done <- (x - y)
	}()
}

func destructActor(a *Actor, call *Call) {
	// Perform cleanup actions here, as needed.
	go func() {
		call.Done <- true
	}()
}

///////////////////////////////
// ACTOR PATTERN TESTS
// (These just help prove that the patterns works)

func TestActorPattern(t *testing.T) {
	a := NewActor()

	if ret := <-a.NoargRet(make(chan bool)); ret {
		t.Errorf("Expected false, got %v", ret)
	}
	a.NoargNoret() // set the state to true
	if ret := <-a.NoargRet(make(chan bool)); !ret {
		t.Errorf("Expected true, got %v", ret)
	}

	// A (poor) example of composing functions:
	first := make(chan int)
	second := make(chan int)
	third := make(chan int)
	go func() {
		// To avoid needing a goroutine here, arguments could be made
		// in to channels as well.
		a.ArgRet(<-first, <-second, third)
	}()
	a.ArgRet(2, 1, second)
	a.ArgRet(4, 3, first)
	if ret := <-third; ret != 0 {
		t.Errorf("Expected 0, got %v", ret)
	}

	<-a.Destruct(make(chan bool))
}
