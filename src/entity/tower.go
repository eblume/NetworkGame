package entity

type Tower struct {
	connect_to        chan *Tower
	disconnect_from   chan *Tower
	destruct          chan bool
	get_num_neighbors chan int
	neighbors         map[*Tower]bool
}

func NewTower() *Tower {
	return &Tower{
		connect_to:        make(chan *Tower),
		disconnect_from:   make(chan *Tower),
		destruct:          make(chan bool),
		get_num_neighbors: make(chan int),
		neighbors:         make(map[*Tower]bool),
	}
}

// Goroutine to run the specified tower
func RunTower(t *Tower) {
	for {

		num_neighbors := len(t.neighbors)

		select {
		case t.destruct <- true:
			for neighbor := range t.neighbors {
				neighbor.disconnect_from <- t
			}
			return

		case other := <-t.connect_to:
			t.neighbors[other] = true

		case other := <-t.disconnect_from:
			delete(t.neighbors, other)

		case t.get_num_neighbors <- num_neighbors:
			// pass

		}
	}
}

// Join the two specified towers.
func JoinTowers(a *Tower, b *Tower) {
	a.connect_to <- b
	b.connect_to <- a
}

// Seperate the two specified towers.

func DisjoinTowers(a *Tower, b *Tower) {
	a.disconnect_from <- b
	b.disconnect_from <- a
}
