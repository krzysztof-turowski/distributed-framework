package test

import (
	"io/ioutil"
	"log"
	"math/rand"
	"testing"

	"github.com/krzysztof-turowski/distributed-framework/byzantine/sync_chor_coan"
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

const (
	ZEROES = 0
	ONES   = 1
	RANDOM = iota
)

func TestChorCoanRandomNoByzantine(t *testing.T) {
	checkLogOutput()
	for iteration := 0; iteration < 500; iteration++ {
		n := 10 + rand.Intn(10)
		t := 0
		testChorCoanIteration(n, t, RANDOM)
	}
}

func TestChorCoanRandomByzantine(t *testing.T) {
	checkLogOutput()
	for iteration := 0; iteration < 500; iteration++ {
		n := 10 + rand.Intn(10)
		t := rand.Intn((n / 3) - 1)
		testChorCoanIteration(n, t, RANDOM)
	}
}

func TestChorCoanAllZeroesAgreeOnZero(t *testing.T) {
	checkLogOutput()
	for iteration := 0; iteration < 500; iteration++ {
		n := 10 + rand.Intn(10)
		t := rand.Intn((n / 3) - 1)
		testChorCoanIteration(n, t, ZEROES)
	}
}

func TestChorCoanAllOnesAgreeOnOne(t *testing.T) {
	checkLogOutput()
	for iteration := 0; iteration < 500; iteration++ {
		n := 10 + rand.Intn(10)
		t := rand.Intn((n / 3) - 1)
		testChorCoanIteration(n, t, ONES)
	}
}

func testChorCoanIteration(n int, t int, initial_values int) {
	V := make([]int, n)
	for i := 0; i < n; i++ {
		if initial_values == RANDOM {
			V[i] = rand.Intn(2)
		} else {
			V[i] = initial_values
		}
	}
	for i := 0; i < t; i++ {
		V[rand.Intn(n)] = -1
	}
	nodes, synchronizer := lib.BuildCompleteGraphWithLoops(n, true, lib.GetGenerator())
	ans, val := sync_chor_coan.Run(nodes, synchronizer, t, V)
	if !ans || (initial_values != RANDOM && val != initial_values) {
		log.Panic("FAILED")
	}
}

func BenchmarkChorCoan(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		n := 10 + rand.Intn(10)
		t := rand.Intn((n / 3) - 1)
		testChorCoanIteration(n, t, RANDOM)
	}
}
