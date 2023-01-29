package test

import (
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/async_hirschberg_sinclair"
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/async_stages_with_feedback"
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/sync_hirschberg_sinclair"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"io/ioutil"
	"log"
	"testing"
)

func TestUndirectedRingSyncHirschbergSinclair(t *testing.T) {
	checkLogOutput()
	sync_hirschberg_sinclair.Run(1000)
}

func BenchmarkUndirectedRingSyncHirschbergSinclair(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_hirschberg_sinclair.Run(1000)
	}
}

func TestUndirectedRingAsyncHirschbergSinclair(t *testing.T) {
	checkLogOutput()
	async_hirschberg_sinclair.Run(100)
}

func BenchmarkUndirectedRingAsyncHirschbergSinclair(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for i := 0; i < b.N; i++ {
		async_hirschberg_sinclair.Run(100)
	}
}

func TestStagesWithFeedback(t *testing.T) {
	checkLogOutput()

	for n := 2; n <= 100; n++ {
		nodes, runner := lib.BuildRing(n)
		async_stages_with_feedback.Run(nodes, runner)
	}
}

func BenchmarkStagesWithFeedback(b *testing.B) {
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		nodes, runner := lib.BuildRing(100)
		async_stages_with_feedback.Run(nodes, runner)
	}
}
