package test

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/krzysztof-turowski/distributed-framework/graphs/ds/sync_kuhn_wattenhofer"
	"github.com/krzysztof-turowski/distributed-framework/graphs/ds/sync_lrg"
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

func TestLRG(t *testing.T) {
	checkLogOutput()
	sync_lrg.Run(600, 0.75)
}

func TestCliqueDeltaKuhnWattenhofer(t *testing.T) {
	vertices, synchronizer := lib.BuildSynchronizedCompleteGraph(20)
	sync_kuhn_wattenhofer.RunWithMaxDegree(vertices, synchronizer, 6)
}

func TestCliqueLocalKuhnWattenhofer(t *testing.T) {
	vertices, synchronizer := lib.BuildSynchronizedCompleteGraph(20)
	sync_kuhn_wattenhofer.Run(vertices, synchronizer, 6)
}

func TestCycleDeltaKuhnWattenhofer(t *testing.T) {
	vertices, synchronizer := lib.BuildSynchronizedRing(101)
	sync_kuhn_wattenhofer.RunWithMaxDegree(vertices, synchronizer, 2)
}

func TestCycleLocalKuhnWattenhofer(t *testing.T) {
	vertices, synchronizer := lib.BuildSynchronizedRing(101)
	sync_kuhn_wattenhofer.Run(vertices, synchronizer, 2)
}

func TestDeltaKuhnWattenhofer(t *testing.T) {
	vertices, synchronizer := lib.BuildSynchronizedRandomGraph(101, 0.05)
	sync_kuhn_wattenhofer.RunWithMaxDegree(vertices, synchronizer, 4)
}

func TestLocalKuhnWattenhofer(t *testing.T) {
	vertices, synchronizer := lib.BuildSynchronizedRandomGraph(101, 0.05)
	sync_kuhn_wattenhofer.Run(vertices, synchronizer, 4)
}

func BenchmarkLRG(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_lrg.Run(600, 0.75)
	}
}

func BenchmarkDeltaKuhnWattenhofer(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		vertices, synchronizer := lib.BuildSynchronizedRandomGraph(600, 0.75)
		sync_kuhn_wattenhofer.RunWithMaxDegree(vertices, synchronizer, 4)
	}
}

func BenchmarkLocalKuhnWattenhofer(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		vertices, synchronizer := lib.BuildSynchronizedRandomGraph(600, 0.75)
		sync_kuhn_wattenhofer.Run(vertices, synchronizer, 4)
	}
}
