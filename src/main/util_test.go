package main

import (
	"testing"
)

func add(a Proto, b Proto) Proto {
	return a.(int) + b.(int)
}

func double(a Proto) Proto {
	return a.(int) * 2
}

func odds(a Proto) bool {
	return a.(int)%2 == 1
}

func TestSendGatherMapReduceFilter(t *testing.T) {
	in := []Proto{0, 1, 2, 3, 4, 5, 6}
	eighteen :=
		Gather(
			Reduce(add,
				Map(double,
					Filter(odds,
						Send(in)))))[0].(int)

	if eighteen != 18 {
		t.Errorf("Expected 18, got %v", eighteen)
	}
}

func TestGatherN(t *testing.T) {
	in := []Proto{0, 1, 2, 3, 4, 5, 6}
	sum := 0
	count := 0
	for tuple := range GatherN(Send(in), 2) {
		count++
		for i := range tuple {
			sum += tuple[i].(int)
		}
	}

	if count != 4 {
		t.Errorf("Expected 4, got %v", count)
	}

	if sum != 21 {
		t.Errorf("Expected 21, got %v", sum)
	}
}

func TestPMap(t *testing.T) {
	in := []Proto{0, 1, 2, 3, 4, 5, 6}
	out := Gather(PMap(double, Send(in)))
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
	out := Gather(PFilter(odds, Send(in)))
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
		t.Error("Expected 9, got %v", sum)
	}

}

func TestMultiplex(t *testing.T) {
	in := []Proto{0, 1, 2, 3, 4, 5, 6}
	result :=
		Gather(
			Reduce(add,
				Multiplex(
					Demultiplex(odds,
						Send(in)))))[0].(int)
	if result != 21 {
		t.Error("Expected 21, got %v", result)
	}
}

func TestParallelMapReduceFilter(t *testing.T) {
	in := []Proto{0, 1, 2, 3, 4, 5, 6}
	eighteen :=
		Gather(
			PReduce(add,
				PMap(double,
					PFilter(odds,
						Send(in)))))[0].(bool)

	if eighteen != false {
		t.Errorf("Expected 18, got %v", eighteen)
	}
}
