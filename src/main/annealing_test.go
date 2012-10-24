package main

// Commented out due to horribleness of testing

// import (
// 	"math/rand"
// 	"strconv"
// 	"testing"
// )

// func TestPathAnnealing(t *testing.T) {
// 	const trials = 5
// 	t.Logf("Testing annealing, which is a nondeterministic algorithm.")
// 	t.Logf("This test may fail without necessarily indicating an error.")
// 	t.Logf("Just rerun the tests to see if it eventually passes - if not, error.")

// 	// Set testing mode for the tower module
// 	TOWER_DEBUG()

// 	// A path 'anneals' if it shows that the path length decreases over 20
// 	// packets. (This is what annealTowers tests.) Because annealing is a
// 	// nondeterministic process, we don't directly test this hypothesis. Instead,
// 	// we test to disprove the null hypothesis, which is that the return value
// 	// of annealTowers is a 50/50 chance. If we can be reasonably sure that
// 	// it's better than 50/50 (in favor of a shorter path - ie. a 'true' return -
// 	// then we consider the test successful.
// 	//
// 	// If we run N such tests, we would expect N/2 to pass (if the null hypoth.
// 	// holds). If we see .8 * N or more pass, we will consider it a success. This is
// 	// chosen arbitrarily.
// 	passed := 0
// 	for i := 0; i < trials; i++ {
// 		if annealTowers(t) {
// 			passed++
// 		}
// 	}

// 	if passed < (0.8 * trials) {
// 		t.Errorf("Only %v of %v annealing trials passed, which is too few.", passed, trials)
// 	}
// }

// func annealTowers(t *testing.T) bool {
// 	// Create a complicated network of 20 towers
// 	towers := createTowers(t, 20)
// 	// Pick a start and stop tower, randomly.
// 	all_rand := rand.Perm(len(towers))
// 	start := towers[all_rand[0]]
// 	stop := towers[all_rand[1]]

// 	// Send 50 packets, noting an overall decrease in trip length

// 	first := countJourney(t, start, stop)
// 	for i := 0; i < 48; i++ {
// 		countJourney(t, start, stop)
// 	}
// 	last := countJourney(t, start, stop)

// 	if first < last {
// 		t.Logf("Failed to anneal path (%d to %d) - this MAY be OK", first, last)
// 	} else {
// 		t.Logf("Annealed path length from %d to %d, hooray!", first, last)
// 	}

// 	// Shutdown
// 	for i := range towers {
// 		<-towers[i].Destruct(make(chan bool))
// 	}

// 	return first >= last
// }

// func createTowers(t *testing.T, count int) []*Tower {
// 	towers := make([]*Tower, count)

// 	// First, create the towers themselves
// 	for i := range towers {
// 		towers[i] = NewTower(strconv.Itoa(i))
// 	}

// 	// Then, add a single 'loop' of connections
// 	all_towers_random := rand.Perm(len(towers))
// 	for i := range all_towers_random {
// 		this := towers[all_towers_random[i]]
// 		next := towers[all_towers_random[(i+1)%len(all_towers_random)]]
// 		<-this.JoinTower(next, make(chan bool))
// 		t.Logf("Joined tower %v and %v", this.name, next.name)
// 	}

// 	// For 50 % of towers, add another connection each
// 	half_towers_random := rand.Perm((len(towers) / 2) + 1)
// 	for i := range half_towers_random {
// 		this := towers[half_towers_random[i]]
// 		next := towers[half_towers_random[(i+1)%len(half_towers_random)]]
// 		<-this.JoinTower(next, make(chan bool))
// 		t.Logf("Joined tower %v and %v", this.name, next.name)

// 	}

// 	// For 10% of towers, add another connection each
// 	dec_towers_random := rand.Perm((len(towers) / 10) + 1)
// 	for i := range dec_towers_random {
// 		this := towers[half_towers_random[i]]
// 		next := towers[half_towers_random[(i+1)%len(dec_towers_random)]]
// 		<-this.JoinTower(next, make(chan bool))
// 		t.Logf("Joined tower %v and %v", this.name, next.name)

// 	}

// 	return towers
// }
