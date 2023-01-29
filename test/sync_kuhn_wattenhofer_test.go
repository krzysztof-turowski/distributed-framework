package test

import (
	"github.com/krzysztof-turowski/distributed-framework/graphs/mds/sync_kuhn_wattenhofer"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"testing"
)

func TestDeltaVariantOnClique(t *testing.T) {
	vertices, synchronizer := lib.BuildSynchronizedCompleteGraph(20)
	sync_kuhn_wattenhofer.RunUsingMaxDegreeInGraph(vertices, synchronizer, 6) // opt =1
}

func TestLocalVariantOnClique(t *testing.T) {
	vertices, synchronizer := lib.BuildSynchronizedCompleteGraph(20)
	sync_kuhn_wattenhofer.Run(vertices, synchronizer, 6) // opt = 1
}

func TestDeltaOnCycle(t *testing.T) {
	vertices, synchronizer := lib.BuildSynchronizedRing(101)
	sync_kuhn_wattenhofer.RunUsingMaxDegreeInGraph(vertices, synchronizer, 2) // opt = 51
}

func TestLocalOnCycle(t *testing.T) {
	vertices, synchronizer := lib.BuildSynchronizedRing(101)
	sync_kuhn_wattenhofer.Run(vertices, synchronizer, 2) // opt = 51
}

func TestDeltaRandom(t *testing.T) {
	vertices, synchronizer := lib.BuildSynchronizedRandomGraph(101, 0.05)
	sync_kuhn_wattenhofer.RunUsingMaxDegreeInGraph(vertices, synchronizer, 4) // opt = ??
}

func TestLocalRandom(t *testing.T) {
	vertices, synchronizer := lib.BuildSynchronizedRandomGraph(101, 0.05)
	sync_kuhn_wattenhofer.Run(vertices, synchronizer, 4) // opt = ??
}
