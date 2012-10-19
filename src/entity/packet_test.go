package entity

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

	a.JoinTower(b)
	b.JoinTower(c)
	c.JoinTower(d)
	d.JoinTower(e)
	b.JoinTower(d)

	testJourney(t, a, e)

	b.DisjoinTower(c)

	//testJourney(t, a, e)

	a.Stop()
	b.Stop()
	c.Stop()
	d.Stop()
	e.Stop()
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
