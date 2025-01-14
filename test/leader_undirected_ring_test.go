package test

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/async_franklin"
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/async_hirschberg_sinclair"
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/async_probabilistic_franklin"
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/async_stages_with_feedback"
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/sync_franklin"
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/sync_franklin_2"
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/sync_higham_przytycka"
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/sync_hirschberg_sinclair"
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/sync_prob_as_far"
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

func TestUndirectedRingAsyncHirschbergSinclair(t *testing.T) {
	checkLogOutput()
	for n := 2; n <= 100; n++ {
		async_hirschberg_sinclair.Run(n)
	}
}

func BenchmarkUndirectedRingAsyncHirschbergSinclair(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for i := 0; i < b.N; i++ {
		async_hirschberg_sinclair.Run(1000)
	}
}

func TestUndirectedRingSyncHirschbergSinclair(t *testing.T) {
	checkLogOutput()
	for n := 2; n <= 100; n++ {
		sync_hirschberg_sinclair.Run(n)
	}
}

func BenchmarkUndirectedRingSyncHirschbergSinclair(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for i := 0; i < b.N; i++ {
		sync_hirschberg_sinclair.Run(1000)
	}
}

func TestUndirectedRingSyncProbAsFar(t *testing.T) {
	checkLogOutput()
	sync_prob_as_far.Run(1000)
}

func BenchmarkUndirectedRingSyncProbAsFar(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_prob_as_far.Run(1000)
	}
}

func TestUndirectedRingSyncFranklin(t *testing.T) {
	checkLogOutput()
	sync_franklin.Run(1000)
}

func TestUndirectedRingSyncFranklin_2(t *testing.T) {
	checkLogOutput()
	sync_franklin_2.Run(1000)
}

func BenchmarkUndirectedRingSyncFranklin(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_franklin.Run(1000)
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

func TestProbabilisticFranklin(t *testing.T) {
	checkLogOutput()
	async_probabilistic_franklin.Run(1000, 3)
}

func BenchmarkProbabilisticFranklin(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		async_probabilistic_franklin.Run(1000, 3)
	}
}

func TestFranklin(t *testing.T) {
	checkLogOutput()

	for n := 2; n <= 100; n++ {
		nodes, runner := lib.BuildRing(n)
		async_franklin.Run(nodes, runner)
	}
}

func BenchmarkFranklin(b *testing.B) {
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		nodes, runner := lib.BuildRing(100)
		async_franklin.Run(nodes, runner)
	}
}

func TestUndirectedRingHighamPrzytycka(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	for n := 2; n <= 100; n++ {
		sync_higham_przytycka.Run(n)
	}
}

func BenchmarkUndirectedRingHighamPrzytycka(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for i := 0; i < b.N; i++ {
		sync_higham_przytycka.Run(100)
	}
}
