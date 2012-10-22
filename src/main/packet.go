package main

type Packet struct {
	journey chan *Tower
	dest    *Tower
}

func NewPacket(dest *Tower, journey chan *Tower) *Packet {
	return &Packet{
		dest:    dest,
		journey: journey,
	}
}
