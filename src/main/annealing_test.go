package main

// Commented out due to horribleness of testing

import (
	"math/rand"
	"strconv"
	"testing"
)

func TestPathAnnealing(t *testing.T) {
	const trials = 20
	// t.Logf("Testing annealing, which is a nondeterministic algorithm.")
	// t.Logf("This test may fail without necessarily indicating an error.")
	// t.Logf("Just rerun the tests to see if it eventually passes - if not, error.")

	// Set testing mode for the tower module
	TOWER_DEBUG()

	// A path 'anneals' if it shows that the path length decreases after being used by
	// packets. (This is what annealTowers tests.) Because annealing is a
	// nondeterministic process, we don't directly test this hypothesis. Instead,
	// we test to disprove the null hypothesis, which is that the return value
	// of annealTowers is a 50/50 chance. If we can be reasonably sure that
	// it's better than 50/50 (in favor of a shorter path - ie. a 'true' return -
	// then we consider the test successful.
	//
	// If we run N such tests, we would expect N/2 to pass (if the null hypoth.
	// holds). If we see .8 * N or more pass, we will consider it a success. This is
	// chosen arbitrarily.
	passed := 0
	pass_chan := make(chan bool)
	for i := 0; i < trials; i++ {
		go func() {
			pass_chan <- annealTowers(t)
		}()
	}
	for i := 0; i < trials; i++ {
		if <-pass_chan {
			passed++
		}
	}

	if passed < (0.8 * trials) {
		t.Errorf("Only %v of %v annealing trials passed, which is too few.", passed, trials)
	}
}

func annealTowers(t *testing.T) bool {
	// Create a complicated network of 20 towers
	towers := createTowers(t, 20)
	// Pick a start and stop tower, randomly.
	all_rand := rand.Perm(len(towers))
	start := towers[all_rand[0]]
	stop := towers[all_rand[1]]

	// Send a packet, record the length
	first := countJourney(t, start, stop)

	// Send a packet from every tower to every other tower.
	sendPath := func() {
		for i := range towers {
			for j := range towers {
				if i == j {
					continue
				}
				countJourney(t, towers[i], towers[j])
			}
		}
	}
	// Do it a few times, to help anneal the network.
	done := make(chan bool)
	for i := 0; i < 3; i++ {
		go func() {
			sendPath()
			done <- true
		}()
	}
	for i := 0; i < 3; i++ {
		<-done
	}

	// Send the original packet again, recording the length
	last := countJourney(t, start, stop)

	// The journey should be smaller or equal to the first!
	if last > first {
		// t.Errorf("Annealing failed, started with %v, ended with %v", first, last)
	}

	// Shutdown
	for i := range towers {
		<-towers[i].Destruct(make(chan bool))
	}

	return first >= last
}

// Create a denseley connected randomized tower net of 'count' towers
func createTowers(t *testing.T, count int) []*Tower {
	towers := make([]*Tower, count)

	// First, create the towers themselves
	for i := range towers {
		towers[i] = NewTower(strconv.Itoa(i))
	}

	// Then, add a single 'loop' of connections
	all_towers_random := rand.Perm(len(towers))
	for i := range all_towers_random {
		this := towers[all_towers_random[i]]
		next := towers[all_towers_random[(i+1)%len(all_towers_random)]]
		<-this.JoinTower(next, make(chan bool))
		// t.Logf("Joined tower %v and %v", this.name, next.name)
	}

	// Then run a few loops of adding some other connections
	connectTowers(towers)
	// connectTowers(towers)
	// connectTowers(towers)

	return towers
}

func connectTowers(towers []*Tower) {
	random_list := rand.Perm(len(towers))
	for i := range random_list {
		a := towers[random_list[i]]
		b := towers[i]
		if a == b {
			continue
		}
		<-a.JoinTower(b, make(chan bool))
	}
}
