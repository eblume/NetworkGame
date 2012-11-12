package main

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestTowerLinking(t *testing.T) {
	towers := createTowers(10)
	for i := range towers {
		if len(towers[i].neighbors.neighbors) != 0 {
			t.Error("Tower had neighbor before it should have.")
		}
	}

	makeCycle(towers, t)
	for i := range towers {
		n := len(towers[i].neighbors.neighbors)
		if n != 2 {
			t.Errorf("Tower %v had %v neighbors, expected 2", towers[i].name, n)
		}
	}

	destroyTowers(towers)
}

// Create `count` towers, totally unconnected.
func createTowers(count int) []*Tower {
	towers := make([]*Tower, count)
	for i := 0; i < count; i++ {
		towers[i] = NewTower(strconv.Itoa(i))
	}
	return towers
}

func destroyTowers(towers []*Tower) {
	for i := range towers {
		towers[i].Destroy()
	}
}

// Create a (random) 'tour' of links through the towers.
// This is not guaranteed to be a NEW cycle, or even to not overlap one.
func makeCycle(towers []*Tower, t *testing.T) {
	walk := rand.Perm(len(towers))
	for i := range walk {
		this := towers[walk[i]]
		next := towers[walk[(i+1)%len(walk)]]
		// We put this in a goroutine just to stress that the function
		// is asynchronous - it's asynch even without the go though.
		go this.JoinTower(next)
	}
	// We must sleep because JoinTower is nonblocking and can
	// become a race condition. Yes this is a kludge.
	time.Sleep(1 * time.Millisecond)
}
