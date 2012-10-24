package main

import (
	"sync"
)

type Packet struct {
	journey  chan *Tower
	dest     *Tower
	trip     []*Tower
	tripLock sync.Mutex
}

func NewPacket(dest *Tower, journey chan *Tower) *Packet {
	return &Packet{
		dest:    dest,
		journey: journey,
		trip:    make([]*Tower, 0, 10),
	}
}

// Record a new hop on the packet's trip.
func (p *Packet) RecordHop(t *Tower) {
	p.tripLock.Lock()
	p.trip = append(p.trip, t)
	p.tripLock.Unlock()
	p.journey <- t
}

// Return a copy of the packet's trip log
func (p *Packet) GetTrip() []*Tower {
	p.tripLock.Lock()
	copy := make([]*Tower, len(p.trip))
	for i := range p.trip {
		copy[i] = p.trip[i]
	}
	p.tripLock.Unlock()
	return copy
}
