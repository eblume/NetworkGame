package main

import (
	"math/rand"
)

const (
	MAX_HELD_PACKETS  int     = 50
	PROB_DELETE_STALE float32 = 0.1
	PROB_PATHFINDER   float32 = 0.1
)

var (
	// This is *LIKE* a constant, except that during testing we want to set it to
	// something completely different.
	SLEEP_PACKET_INTERVAL int = 500
)

func TOWER_DEBUG() {
	SLEEP_PACKET_INTERVAL = 1
}

type towerDistance struct {
	tower    *Tower
	distance int
}

type Tower struct {
	Method_destruct        chan *Call
	Method_getnumneighbors chan *Call
	Method_jointower       chan *Call
	Method_disjointower    chan *Call
	Method_handlepacket    chan *Call
	// Internals
	neighbors   map[*Tower]bool
	linktower   chan *Tower
	unlinktower chan *Tower
	name        string
	packets     chan *Packet
	cache       map[*Tower]*towerDistance
}

/////////////// METHODS

// Create a new Tower and start it's handling goroutine
func NewTower(name string) *Tower {
	t := &Tower{
		Method_destruct:        make(chan *Call),
		Method_getnumneighbors: make(chan *Call),
		Method_jointower:       make(chan *Call),
		Method_disjointower:    make(chan *Call),
		Method_handlepacket:    make(chan *Call),
		// Internals
		neighbors:   make(map[*Tower]bool),
		linktower:   make(chan *Tower),
		unlinktower: make(chan *Tower),
		name:        name,
		packets:     make(chan *Packet, MAX_HELD_PACKETS),
		cache:       make(map[*Tower]*towerDistance),
	}
	go run(t)
	return t
}

func (t *Tower) Destruct(done chan bool) chan bool {
	go func() {
		recv := make(chan interface{})
		t.Method_destruct <- &Call{
			Done: recv,
		}
		done <- (<-recv).(bool)
	}()
	return done
}

func (t *Tower) GetNumNeighbors(done chan int) chan int {
	go func() {
		recv := make(chan interface{})
		t.Method_getnumneighbors <- &Call{
			Done: recv,
		}
		done <- (<-recv).(int)
	}()
	return done
}

func (t *Tower) JoinTower(other *Tower, done chan bool) chan bool {
	go func() {
		recv := make(chan interface{})
		t.Method_jointower <- &Call{
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
		t.Method_disjointower <- &Call{
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
		t.Method_handlepacket <- &Call{
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
		case call := <-t.Method_getnumneighbors:
			getnumneighbors(t, call)
		case call := <-t.Method_jointower:
			jointower(t, call)
		case call := <-t.Method_disjointower:
			disjointower(t, call)
		case call := <-t.Method_handlepacket:
			handlepacket(t, call)
		case call := <-t.Method_destruct:
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
	// If the packet has reached it's dest, stop it!
	if t == p.dest {
		close(p.journey)
		return
	}
	recordPacketTrail(t, p)
	if _, exists := t.neighbors[p.dest]; exists {
		go sendPacket(p.dest, p)
		return
	}

	// With some probability, disregard the cache and travel
	// to some other neighbor than the last one randomly.
	trip := p.GetTrip()
	neighbors := make([]*Tower, len(t.neighbors))
	i := 0
	for k := range t.neighbors {
		neighbors[i] = k
		i++
	}
	if len(t.neighbors) > 1 && rand.Float32() <= PROB_PATHFINDER {
		// Chose a random neighbor other than the one we were at last.
		for j := range rand.Perm(len(neighbors)) {
			if len(trip) < 2 || (neighbors[j] == trip[len(trip)-2]) {
				continue
			}
			go sendPacket(neighbors[j], p)
			return
		}
	}

	// If the destination is in the travel cache, just take that!
	if next, cache_ok := t.cache[p.dest]; cache_ok {
		if _, link_ok := t.neighbors[next.tower]; link_ok {
			go sendPacket(next.tower, p)
			return
		} else {
			// Stale cache entry, delete it!
			delete(t.cache, p.dest)
			// We could also delete such entries when unlinking, but this
			// feels more correct somehow.
		}
	}

	// Finally, just choose a place to go at random!
	for j := range rand.Perm(len(neighbors)) {
		go sendPacket(neighbors[j], p)
		return
	}
}

// GOROUTINE to send the packet.
func sendPacket(next *Tower, p *Packet) {
	next.packets <- p
	p.RecordHop(next)
}

// Return true iff 'other' is 't' or a neighbor of 't'.
func isMeOrNeighbor(t *Tower, other *Tower) bool {
	if t == other {
		return true
	}

	for neighbor := range t.neighbors {
		if other == neighbor {
			return true
		}
	}

	return false
}

// Use the packet's trail as clues to populate the cache
func recordPacketTrail(t *Tower, p *Packet) {
	trip := p.GetTrip()
	if len(trip) < 2 {
		return // nothing to record
	}
	neighbor := trip[len(trip)-2]
	// neighbor might have been deleted, but we'll let it get in to
	// the cache anyway and handle all such failures in one place.
	saw_myself := false
	for i := range trip {
		if isMeOrNeighbor(t, trip[i]) {
			continue
		}

		hop := trip[i] // SO close to writing trip[hop] - damn
		trip_distance := len(trip) - i - 1
		other, exists := t.cache[hop]
		if !exists || (trip_distance < other.distance) {
			t.cache[hop] = &towerDistance{neighbor, trip_distance}
		}

		if hop == t {
			saw_myself = true
		}
	}

	if _, ok := t.cache[p.dest]; ok && saw_myself {
		// We've seen this packet before, yet we NOW think we could have routed
		// this packet. Maybe the route is stale, or maybe we routed this packet
		// last time before we new about the current route, or maybe we just got
		// really unlucky with our 'pathfinder' routings.
		//
		// To avoid adding complexity, we'll just delete this route with some
		// probability. (see PROB_DELETE_STALE) The upshot is that this is simple
		// and will eventually converge correctly. The downside is that it will be
		// not uncommon with deleted towers to see packets thrash around for a while
		// until this code path gets executed.
		if rand.Float32() <= PROB_DELETE_STALE {
			delete(t.cache, p.dest)
		}
	}
}

func handlepacket(t *Tower, c *Call) {
	packet := c.Args[0].(*Packet)
	go func() {
		t.packets <- packet
		packet.RecordHop(t)
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
		go func(neighbor *Tower) {
			neighbor.unlinktower <- t
		}(neighbor)
	}
	go func() {
		c.Done <- true
	}()
}
