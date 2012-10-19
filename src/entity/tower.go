package entity

type Tower struct {
    connect_to        chan *Tower
    disconnect_from   chan *Tower
    destruct          chan bool
    get_num_neighbors chan chan int
    neighbors         map[*Tower]bool
}

func NewTower() *Tower {
    t := &Tower{
        connect_to:        make(chan *Tower),
        disconnect_from:   make(chan *Tower),
        destruct:          make(chan bool),
        get_num_neighbors: make(chan chan int),
        neighbors:         make(map[*Tower]bool),
    }
    go runTower(t)
    return t
}

func (t *Tower) GetNumNeighbors() int {
    recv := make(chan int)
    t.get_num_neighbors <- recv
    return <- recv
}

func (t *Tower) JoinTower(other *Tower) {
    t.connect_to <- other
    other.connect_to <- t
}

func (t *Tower) DisjoinTower(other *Tower) {
    t.disconnect_from <- other
    other.disconnect_from <- t
}

func (t *Tower) Stop() {
    <- t.destruct
}


// INTERNALS

func runTower(t *Tower) {
    for {
        select {
        case recv := <- t.get_num_neighbors:
            recv <- len(t.neighbors)            

        case other := <- t.connect_to:
            t.neighbors[other] = true

        case other := <- t.disconnect_from:
            delete(t.neighbors, other)

        case t.destruct <- true:
            destruct(t)
            return
        }
    }
}

func destruct(t *Tower) {
    for other := range t.neighbors {
        other.disconnect_from <- t
    }
}

