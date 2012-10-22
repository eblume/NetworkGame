package main

import (
	"math/rand"
)

const (
	MAX_HELD_PACKETS      int = 50
	SLEEP_PACKET_INTERVAL int = 1500
)

type Tower struct {
	method_destruct        chan *Call
	method_getnumneighbors chan *Call
	method_jointower       chan *Call
	method_disjointower    chan *Call
	method_handlepacket    chan *Call
	// Internals
	neighbors   map[*Tower]bool
	linktower   chan *Tower
	unlinktower chan *Tower
	name        string
	packets     chan *Packet
}

/////////////// METHODS

// Create a new Tower and start it's handling goroutine
func NewTower(name string) *Tower {
	t := &Tower{
		method_destruct:        make(chan *Call),
		method_getnumneighbors: make(chan *Call),
		method_jointower:       make(chan *Call),
		method_disjointower:    make(chan *Call),
		method_handlepacket:    make(chan *Call),
		// Internals
		neighbors:   make(map[*Tower]bool),
		linktower:   make(chan *Tower),
		unlinktower: make(chan *Tower),
		name:        name,
		packets:     make(chan *Packet, MAX_HELD_PACKETS),
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

// Place a brand-new packet on the system, to be handed around.
func (t *Tower) HandlePacket(p *Packet) {
	go func() {
		t.method_handlepacket <- &Call{
			Args: []interface{}{p},
		}
	}()
}

/////////////// HANDLERS

func run(t *Tower) {
	close_packet_ivl := make(chan bool)
	packet_interval := Interval(SLEEP_PACKET_INTERVAL, close_packet_ivl)
	for {
		select {
		case call := <-t.method_getnumneighbors:
			getnumneighbors(t, call)
		case call := <-t.method_jointower:
			jointower(t, call)
		case call := <-t.method_disjointower:
			disjointower(t, call)
		case call := <-t.method_handlepacket:
			handlepacket(t, call)
		case call := <-t.method_destruct:
			close_packet_ivl <- true // close the interval
			destruct(t, call)
			return
		// Internals
		case other := <-t.linktower:
			t.neighbors[other] = true
		case other := <-t.unlinktower:
			delete(t.neighbors, other)
		case <-packet_interval:
			// Try and handle a packet
			select {
			case packet := <-t.packets:
				processPacket(t, packet)
			default:
			}
		}
	}
}

func processPacket(t *Tower, p *Packet) {
	// In the future, do something smart and clever.
	// For now, send it somewhere randomly.
	//
	// Make a copy of the neighbors so we can hand it off.
	var neighbors map[*Tower]bool
	for k, v := range t.neighbors {
		neighbors[k] = v
	}
	go func() {
		// Shuffle the neighbors up
		others := make([]*Tower, len(t.neighbors))
		random := rand.Perm(len(others))
		i := 0
		for other := range t.neighbors {
			others[random[i]] = other
			i++
		}
		// Let's just send it to 'others[0]', no more checking/shuffling!
		next := others[0]
		next.packets <- p
		p.journey <- next
	}()
}

func handlepacket(t *Tower, c *Call) {
	packet := c.Args[0].(*Packet)
	go func() {
		t.packets <- packet
		packet.journey <- t
	}()
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
	for neighbor := range t.neighbors {
		go func() {
			neighbor.unlinktower <- t
		}()
	}
	go func() {
		c.Done <- true
	}()
}
