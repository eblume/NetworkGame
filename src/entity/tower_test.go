package entity

import (
	"testing"
)

func TestTowerConnections(t *testing.T) {

	// Create three towers
	a := NewTower()
	b := NewTower()
	c := NewTower()
	towers := []*Tower{a, b, c}

	// Start up the towers
	go RunTower(a)
	go RunTower(b)
	go RunTower(c)

	// Verify none have any neighbors:
	for tower := range towers {
		if num := <-towers[tower].get_num_neighbors; num != 0 {
			t.Error("Tower had neighbors before adding any.")
		}
	}

	// Add a cycle a <-> b <-> c <-> a  (also means 'connect themn all')
	JoinTowers(a, b)
	JoinTowers(b, c)
	JoinTowers(c, a)

	// Verify that each has 1 neighbor
	for tower := range towers {
		if num := <-towers[tower].get_num_neighbors; num != 2 {
			t.Error("Tower failed to add a neighbor correctly.")
		}
	}

	// Break the chain: a <-> b <-> c
	DisjoinTowers(c, a)

	// Verify that b has 2 neighbors, but a and c have 1
	if num := <-a.get_num_neighbors; num != 1 {
		t.Error("Tower 'a' didn't disconnect from 'c'")
	}
	if num := <-c.get_num_neighbors; num != 1 {
		t.Error("Tower 'c' didn't disconnect from 'a'")
	}
	if num := <-b.get_num_neighbors; num != 2 {
		t.Error("Tower 'b' disconnected but shouldn't have")
	}

}
