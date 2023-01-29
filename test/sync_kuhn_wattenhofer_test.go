package sync_kuhn_wattenhofer

import (
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"testing"
)

func TestDeltaVariantOnClique(t *testing.T) {
	vertices, synchronizer := lib.BuildSynchronizedCompleteGraph(20)
	RunUsingMaxDegreeInGraph(vertices, synchronizer, 6) // opt =1
}

func TestLocalVariantOnClique(t *testing.T) {
	vertices, synchronizer := lib.BuildSynchronizedCompleteGraph(20)
	Run(vertices, synchronizer, 6) // opt = 1
}

func TestDeltaOnCycle(t *testing.T) {
	vertices, synchronizer := lib.BuildSynchronizedRing(101)
	RunUsingMaxDegreeInGraph(vertices, synchronizer, 2) // opt = 51
}

func TestLocalOnCycle(t *testing.T) {
	vertices, synchronizer := lib.BuildSynchronizedRing(101)
	Run(vertices, synchronizer, 2) // opt = 51
}

func TestDeltaRandom(t *testing.T) {
	vertices, synchronizer := lib.BuildSynchronizedRandomGraph(101, 0.05)
	RunUsingMaxDegreeInGraph(vertices, synchronizer, 4) // opt = ??
}

func TestLocalRandom(t *testing.T) {
	vertices, synchronizer := lib.BuildSynchronizedRandomGraph(101, 0.05)
	Run(vertices, synchronizer, 4) // opt = ??
}
