package entity

import (
	"math/rand"
	"time"
)

const (
	MAX_HELD_PACKETS int = 50
	PACKET_WAIT_TIME int = 1000
)

type Tower struct {
	connect_to        chan *Tower
	disconnect_from   chan *Tower
	destruct          chan chan bool
	get_num_neighbors chan chan int
	take_packet       chan *Packet

	// Internal members (not safe for external access!)
	neighbors map[*Tower]bool
	name      string
}

func NewTower(name string) *Tower {
	t := &Tower{
		connect_to:        make(chan *Tower),
		disconnect_from:   make(chan *Tower, 1),
		destruct:          make(chan chan bool),
		get_num_neighbors: make(chan chan int),
		take_packet:       make(chan *Packet, MAX_HELD_PACKETS),
		neighbors:         make(map[*Tower]bool),
		name:              name,
	}
	go runTower(t)
	return t
}

func (t *Tower) GetNumNeighbors() chan int {
	recv := make(chan int)
	t.get_num_neighbors <- recv
	return recv
}

func (t *Tower) JoinTower(other *Tower) chan bool {
	done := make(chan bool, 1)
	go func() {
		t.connect_to <- other
		other.connect_to <- t
		done <- true
	}()
	return done
}

func (t *Tower) DisjoinTower(other *Tower) chan bool {
	done := make(chan bool, 1)
	go func() {
		t.disconnect_from <- other
		other.disconnect_from <- t
		done <- true
	}()
	return done
}

func (t *Tower) Stop() {
	done := make(chan bool, 1)
	t.destruct <- done
	return done
}

// Take a newly minted packet and put it on the 'grid'.
// This does not block - instead, monitor p.journey for updates.
func (t *Tower) HandlePacket(p *Packet) {
	go func() {
		t.take_packet <- p
	}()
}

// INTERNALS

// "Instant" queue - process events that should happen ASAP
func runTower(t *Tower) {
	stop_packets := make(chan bool)
	go runTowerPackets(t, stop_packets)
	for {
		select {
		case recv := <-t.get_num_neighbors:
			recv <- len(t.neighbors)

		case other := <-t.connect_to:
			t.neighbors[other] = true

		case other := <-t.disconnect_from:
			delete(t.neighbors, other)

		case done := <-t.destruct:
			stop_packets <- true
			go destruct(t, done)
			return
		}
	}
}

// Packet-processing queue, which will occur at regular intervals
func runTowerPackets(t *Tower, stop chan bool) {
	for {
		select {
		case <-stop:
			// TODO - what do we do with the packets here?
			return
		case p := <-t.take_packet:
			p.journey <- t
			if p.dest == t {
				close(p.journey)
				// Normally we'd also now 'consume' the packet, but that's
				// going to be added later.
			} else {
				found := false
				for next := range getRoute(t, p) {
					timeout := makeTimeout(1500)
					select {
					case next.take_packet <- p:
						found = true
						break
					case <-timeout:
						continue
					}
				}
				if !found {
					// TODO - We tried all the possible ways of sending the
					// packet, what do we want to do now? Re-queue it? To
					// prevent network starvation, for now we're going to just
					// kill the packet.
					close(p.journey)
				}
			}
			time.Sleep(time.Duration(PACKET_WAIT_TIME) * time.Millisecond)
		}
	}
}

// Return a channel which will generate the neighbors which
// move the packet towards it's destination.
func getRoute(t *Tower, p *Packet) chan *Tower {
	// TODO - implement an ACTUAL search here.
	// For now, it's just a random jump to one of the neighbors.
	// btw - we use channels because we eventually want to distribute
	// this search across the grid in a parallel way, with cacheing.
	recv := make(chan *Tower)
	go func() {
		others := make([]*Tower, len(t.neighbors))
		i := 0
		for other := range t.neighbors {
			others[i] = other
			i++
		}
		for j := range rand.Perm(len(others)) {
			recv <- others[j]
		}
	}()
	return recv
}

func destruct(t *Tower, done chan bool) {
	for other := range t.neighbors {
		other.disconnect_from <- t
	}
	done <- true
}

// Returns a channel that will receive an event
// sometime after 'msecs' milliseconds have passed.
func makeTimeout(msecs int) chan bool {
	to := make(chan bool, 1)
	go func() {
		time.Sleep(time.Duration(msecs) * time.Millisecond)
		to <- true
	}()
	return to
}
