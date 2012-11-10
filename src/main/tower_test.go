package main

import (
	"rand"
	"strconv"
	"testing"
)

func TestTowerLinking(t *testing.T) {
	towers := createTowers(10)
	for i := range towers {
		if len(towers[i].neighbors) != 0 {
			t.Error("Tower had neighbor before it should have.")
		}
	}

	makeCycle(towers)
	for i := range towers {
		if len(towers[i].neighbors) != 2 {
			t.Error("Tower did not have as many neighbors as it should have.")
		}
	}
}

// Create `count` towers, totally unconnected.
func createTowers(count int) []*Tower {
	towers := make([]*Tower, count)
	for i := range count {
		towers[i] = NewTower(strconv.Atoi(i))
	}
	return towers
}

// Create a (random) 'tour' of links through the towers.
// This is not guaranteed to be a NEW cycle, or even to not overlap one.
func makeCycle(towers []*Tower) {
	walk := rand.Perm(len(towers))
	for i := range walk {
		this := towers[walk[i]]
		next := towers[walk[(i+1)%len(walk)]]
		this.JoinTower(next)
	}
}
