package main

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

// For the cache map[*Tower]*towerDistance -> the distance to the key tower,
// along with the neighboring tower to use to get there.
type towerDistance struct {
	via      *Tower
	distance int
}

type Tower struct {
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
		neighbors:   make(map[*Tower]bool),
		linktower:   make(chan *Tower),
		unlinktower: make(chan *Tower),
		name:        name,
		packets:     make(chan *Packet, MAX_HELD_PACKETS),
		cache:       make(map[*Tower]*towerDistance),
	}
	return t
}

func (t *Tower) Destruct(done chan bool) chan bool {
	return nil
}

func (t *Tower) JoinTower(other *Tower, done chan bool) chan bool {
	return nil
}

func (t *Tower) DisjoinTower(other *Tower, done chan bool) chan bool {
	return nil
}
