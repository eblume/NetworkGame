package main

import (
	"testing"
)

func TestTowerConnections(t *testing.T) {

	// Create three towers
	a := NewTower("a")
	b := NewTower("b")
	c := NewTower("c")
	towers := []*Tower{a, b, c}

	// Verify none have any neighbors:
	for tower := range towers {
		recv := make(chan int)
		towers[tower].GetNumNeighbors(recv)
		if num := <-recv; num != 0 {
			t.Errorf("Tower %v had neighbors before adding any.", towers[tower].name)
		}
	}

	// Add a cycle a <-> b <-> c <-> a  (also means "connect them all")
	finished := make(chan bool)
	<-a.JoinTower(b, finished)
	<-b.JoinTower(c, finished)
	<-c.JoinTower(a, finished)

	// Verify that each has 1 neighbor
	for tower := range towers {
		if num := <-towers[tower].GetNumNeighbors(make(chan int)); num != 2 {
			t.Errorf("Tower %v failed to add a neighbor correctly.", towers[tower].name)
		}
	}

	// Break the chain: a <-> b <-> c
	<-a.DisjoinTower(c, finished)

	// Verify that b has 2 neighbors, but a and c have 1
	if num := <-a.GetNumNeighbors(make(chan int)); 1 != num {
		t.Error("Tower a didn't disconnect from c")
	}
	if num := <-c.GetNumNeighbors(make(chan int)); 1 != num {
		t.Error("Tower c didn't disconnect from a")
	}
	if num := <-b.GetNumNeighbors(make(chan int)); 2 != num {
		t.Error("Tower b disconnected but shouldn't have")
	}

	// Shut down the towers
	for tower := range towers {
		<-towers[tower].Destruct(make(chan bool))
	}

}
