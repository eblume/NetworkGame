package main

import (
	"testing"
)

func TestPacket(t *testing.T) {

	// Create a network of towers,
	// a <-> b <-> c <-> d <-> e
	//        ^----------^

	a := NewTower("a")
	b := NewTower("b")
	c := NewTower("c")
	d := NewTower("d")
	e := NewTower("e")

	<-a.JoinTower(b, make(chan bool))
	<-b.JoinTower(c, make(chan bool))
	<-c.JoinTower(d, make(chan bool))
	<-d.JoinTower(e, make(chan bool))
	<-b.JoinTower(d, make(chan bool))

	testJourney(t, a, e)

	// Make the new network,
	// a <-> b     c <-> d <-> e
	//       ^-----------^
	<-b.DisjoinTower(c, make(chan bool))

	testJourney(t, a, e)

	<-a.Destruct(make(chan bool))
	<-b.Destruct(make(chan bool))
	<-c.Destruct(make(chan bool))
	<-d.Destruct(make(chan bool))
	<-e.Destruct(make(chan bool))
}

func testJourney(t *testing.T, start *Tower, stop *Tower) {

	journey := make(chan *Tower)
	p := NewPacket(stop, journey)
	start.HandlePacket(p)

	last := start
	for hop := range journey {
		t.Logf("Jump from %v to %v", last.name, hop.name)
		last = hop
	}

	if last != stop {
		t.Errorf("Packet finished at %v instead of %v", last.name, stop.name)
	}

}
