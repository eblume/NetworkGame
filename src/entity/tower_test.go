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

	// Verify none have any neighbors:
	for tower := range towers {
		if num := towers[tower].GetNumNeighbors(); num != 0 {
			t.Error("Tower had neighbors before adding any.")
		}
	}

	// Add a cycle a <-> b <-> c <-> a  (also means 'connect themn all')
	a.JoinTower(b)
	b.JoinTower(c)
	c.JoinTower(a)

	// Verify that each has 1 neighbor
	for tower := range towers {
		if num := towers[tower].GetNumNeighbors(); num != 2 {
			t.Error("Tower failed to add a neighbor correctly.")
		}
	}

	// Break the chain: a <-> b <-> c
	a.DisjoinTower(c)

	// Verify that b has 2 neighbors, but a and c have 1
	if num := a.GetNumNeighbors(); num != 1 {
		t.Error("Tower 'a' didn't disconnect from 'c'")
	}
	if num := c.GetNumNeighbors(); num != 1 {
		t.Error("Tower 'c' didn't disconnect from 'a'")
	}
	if num := b.GetNumNeighbors(); num != 2 {
		t.Error("Tower 'b' disconnected but shouldn't have")
	}

	// Shut down the towers
	for tower := range towers {
		towers[tower].Stop()
	}

}
