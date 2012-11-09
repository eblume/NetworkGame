package main

import (
	"testing"
)

func add_reduce(a Proto, b Proto) Proto {
	return a.(int) + b.(int)
}

func double_map(a Proto) Proto {
	return a.(int) * 2
}

func filt_odd(a Proto) bool {
	return a.(int)%2 == 1
}

func TestSendGatherMapReduceFilter(t *testing.T) {
	in := []Proto{0, 1, 2, 3, 4, 5, 6}
	eighteen :=
		Gather(
			Reduce(add_reduce,
				Map(double_map,
					Filter(filt_odd,
						Send(in)))))[0].(int)

	if eighteen != 18 {
		t.Errorf("Expected 18, got %v", eighteen)
	}
}

func TestPMap(t *testing.T) {
	in := []Proto{0, 1, 2, 3, 4, 5, 6}
	out := Gather(PMap(double_map, Send(in)))
	count := 0
	sum := 0
	for i := range out {
		sum += out[i].(int)
		count++
	}

	if count != 7 {
		t.Errorf("Expected 7, got %v", count)
	}

	if sum != 42 {
		t.Errorf("Expected 42, got %v", sum)
	}
}

func TestPFilter(t *testing.T) {
	in := []Proto{0, 1, 2, 3, 4, 5, 6}
	out := Gather(PFilter(filt_odd, Send(in)))
	count := 0
	sum := 0
	for i := range out {
		sum += out[i].(int)
		count++
	}

	if count != 3 {
		t.Errorf("Expected 3, got %v", count)
	}

	if sum != 9 {
		t.Errorf("Expected 9, got %v", sum)
	}

}

func TestMultiplex(t *testing.T) {
	in := []Proto{0, 1, 2, 3, 4, 5, 6}
	// Double the odds, then add them all up.
	odd, even := Demultiplex(filt_odd, Send(in))
	combined := Multiplex(Map(double_map, odd), even)
	result := Gather(Reduce(add_reduce,combined))[0].(int)
	if result != 30 {
		t.Errorf("Expected 30, got %v", result)
	}
}