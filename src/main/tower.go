package main

import (
	"sync"
)

type Tower struct {
	name      string
	neighbors map[*Tower]bool
}

func NewTower(name string) *Tower {
	t := &Tower{
		name:      name,
		neighbors: make(map[*Tower]bool),
	}
	return t
}

func (t *Tower) JoinTower(other *Tower, done chan bool) chan bool {
	return nil
}

func (t *Tower) DisjoinTower(other *Tower, done chan bool) chan bool {
	return nil
}
