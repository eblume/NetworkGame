package main

type Tower struct {
	method_destruct        chan *Call
	method_getnumneighbors chan *Call
	method_jointower       chan *Call
	method_disjointower    chan *Call
	// Internals
	neighbors   map[*Tower]bool
	linktower   chan *Tower
	unlinktower chan *Tower
	name        string
}

/////////////// METHODS

// Create a new Tower and start it's handling goroutine
func NewTower(name string) *Tower {
	t := &Tower{
		method_destruct:        make(chan *Call),
		method_getnumneighbors: make(chan *Call),
		method_jointower:       make(chan *Call),
		method_disjointower:    make(chan *Call),
		// Internals
		neighbors:   make(map[*Tower]bool),
		linktower:   make(chan *Tower),
		unlinktower: make(chan *Tower),
		name:        name,
	}
	go run(t)
	return t
}

func (t *Tower) Destruct(done chan bool) chan bool {
	go func() {
		recv := make(chan interface{})
		t.method_destruct <- &Call{
			Done: recv,
		}
		done <- (<-recv).(bool)
	}()
	return done
}

func (t *Tower) GetNumNeighbors(done chan int) chan int {
	go func() {
		recv := make(chan interface{})
		t.method_getnumneighbors <- &Call{
			Done: recv,
		}
		done <- (<-recv).(int)
	}()
	return done
}

func (t *Tower) JoinTower(other *Tower, done chan bool) chan bool {
	go func() {
		recv := make(chan interface{})
		t.method_jointower <- &Call{
			Args: []interface{}{other},
			Done: recv,
		}
		done <- (<-recv).(bool)
	}()
	return done
}

func (t *Tower) DisjoinTower(other *Tower, done chan bool) chan bool {
	go func() {
		recv := make(chan interface{})
		t.method_disjointower <- &Call{
			Args: []interface{}{other},
			Done: recv,
		}
		done <- (<-recv).(bool)
	}()
	return done
}

/////////////// HANDLERS

func run(t *Tower) {
	for {
		select {
		case call := <-t.method_getnumneighbors:
			getnumneighbors(t, call)
		case call := <-t.method_jointower:
			jointower(t, call)
		case call := <-t.method_disjointower:
			disjointower(t, call)
		case call := <-t.method_destruct:
			destruct(t, call)
			return
		// Internals
		case other := <-t.linktower:
			t.neighbors[other] = true
		case other := <-t.unlinktower:
			delete(t.neighbors, other)
		}
	}
}

func jointower(t *Tower, c *Call) {
	other := c.Args[0].(*Tower)
	go func() {
		t.linktower <- other
		other.linktower <- t
		c.Done <- true
	}()
}

func disjointower(t *Tower, c *Call) {
	other := c.Args[0].(*Tower)
	go func() {
		t.unlinktower <- other
		other.unlinktower <- t
		c.Done <- true
	}()
}

func getnumneighbors(t *Tower, c *Call) {
	num := len(t.neighbors)
	go func() {
		c.Done <- num
	}()
}

func destruct(t *Tower, c *Call) {
	go func() {
		c.Done <- true
	}()
}
