package test

import (
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring"
	"github.com/krzysztof-turowski/distributed-framework/leader/ring"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"io/ioutil"
	"log"
	"testing"
)

func TestUndirectedRingHirschbergSinclair(t *testing.T) {
	checkLogOutput()
	undirected_ring.RunHirschbergSinclair(1000)
}

func BenchmarkUndirectedRingHirschbergSinclair(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		undirected_ring.RunHirschbergSinclair(1000)
	}
}

func TestStagesWithFeedback(t *testing.T) {
	checkLogOutput()

	for n := 2; n <= 100; n++ {
		nodes, runner := lib.BuildRing(n)
		ring.RunStagesWithFeedback(nodes, runner)
	}
}

func BenchmarkStagesWithFeedback(b *testing.B) {
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		nodes, runner := lib.BuildRing(100)
		ring.RunStagesWithFeedback(nodes, runner)
	}
}
