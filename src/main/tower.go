package main

type TowerNeighbors struct {
	neighbors     map[*Tower]bool
	join_tower    chan *Tower
	disjoin_tower chan *Tower
	stop          chan bool
}

type Tower struct {
	name      string
	neighbors TowerNeighbors
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
