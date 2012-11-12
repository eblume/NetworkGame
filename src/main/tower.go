package main

const (
	PACKET_QUEUE_SIZE = 5
)

type TowerNeighbors struct {
	neighbors     map[*Tower]bool
	join_tower    chan *Tower
	disjoin_tower chan *Tower
	stop          chan bool
}

type TowerPackets struct {
	packets       chan *Packet
	interval      chan bool
	stop_interval chan bool
	stop          chan bool
}

type Tower struct {
	name      string
	neighbors TowerNeighbors
	packets   TowerPackets
}

func NewTower(name string) *Tower {
	t := &Tower{
		name: name,
		neighbors: TowerNeighbors{
			neighbors:     make(map[*Tower]bool),
			join_tower:    make(chan *Tower),
			disjoin_tower: make(chan *Tower),
			stop:          make(chan bool, 1),
		},
		packets: TowerPackets{
			packets:       make(chan *Packet, PACKET_QUEUE_SIZE),
			interval:      make(chan bool),
			stop_interval: make(chan bool),
			stop:          make(chan bool),
		},
	}
	go monitor_neighbors(t)
	return t
}

func monitor_neighbors(t *Tower) {
	for {
		select {
		case <-t.neighbors.stop:
			return
		case other := <-t.neighbors.join_tower:
			t.neighbors.neighbors[other] = true
		case other := <-t.neighbors.disjoin_tower:
			delete(t.neighbors.neighbors, other)
		}
	}
}

// Simple utility function to spontaneously 'create' (and begin
// routing) a given packet. None of the usual packet creation logic
// is run, it just immediatly begins routing the packet.
func (t *Tower) InjectPacket(p *Packet) {

}

func (t *Tower) Destroy() {
	t.neighbors.stop <- true // Nonblocking for the first call
}

func (t *Tower) JoinTower(other *Tower) {
	t.neighbors.join_tower <- other
	other.neighbors.join_tower <- t
}

func (t *Tower) DisjoinTower(other *Tower) {
	t.neighbors.disjoin_tower <- other
	other.neighbors.disjoin_tower <- t
}
