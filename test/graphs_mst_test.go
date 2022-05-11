package test

import (
	"github.com/krzysztof-turowski/distributed-framework/graphs/mst/sync_ghs"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"io/ioutil"
	"log"
	"testing"
)

func TestRunSynchronizedGHSTriangle(t *testing.T) {
	checkLogOutput()
	adjacencyList := [][][2]int{
		{{2, 10}, {3, 10}},
		{{1, 10}, {3, 20}},
		{{1, 10}, {2, 20}},
	}
	graph, synchronizer := lib.BuildSynchronizedWeightedGraphFromAdjacencyList(adjacencyList, lib.GetGenerator())
	sync_ghs.RunSynchronizedGHS(graph, synchronizer)
}

func TestRunSynchronizedGHSSimplePath4(t *testing.T) {
	checkLogOutput()
	adjacencyList := [][][2]int{
		{{2, 10}},
		{{1, 10}, {3, 20}},
		{{2, 20}, {4, 10}},
		{{3, 10}},
	}
	graph, synchronizer := lib.BuildSynchronizedWeightedGraphFromAdjacencyList(adjacencyList, lib.GetGenerator())
	sync_ghs.RunSynchronizedGHS(graph, synchronizer)
}

func TestRunSynchronizedGHSSimplePath8(t *testing.T) {
	checkLogOutput()
	adjacencyList := [][][2]int{
		{{2, 10}},
		{{1, 10}, {3, 20}},
		{{2, 20}, {4, 10}},
		{{3, 10}, {5, 30}},
		{{4, 30}, {6, 10}},
		{{5, 10}, {7, 20}},
		{{6, 20}, {8, 10}},
		{{7, 10}},
	}
	graph, synchronizer := lib.BuildSynchronizedWeightedGraphFromAdjacencyList(adjacencyList, lib.GetGenerator())
	sync_ghs.RunSynchronizedGHS(graph, synchronizer)
}

func TestRunSynchronizedGHSWikipediaExample(t *testing.T) {
	checkLogOutput()
	adjacencyList := [][][2]int{
		{{2, 4}, {3, 1}, {4, 4}},
		{{1, 4}, {3, 3}, {5, 10}, {10, 18}},
		{{1, 1}, {2, 3}, {4, 5}, {5, 9}},
		{{1, 4}, {3, 5}, {5, 7}, {6, 9}, {7, 9}},
		{{2, 10}, {3, 9}, {4, 7}, {7, 8}, {8, 9}, {10, 8}},
		{{4, 9}, {7, 2}, {8, 4}, {9, 6}},
		{{4, 9}, {5, 8}, {6, 2}, {8, 2}},
		{{5, 9}, {6, 4}, {7, 2}, {9, 3}, {10, 9}},
		{{6, 6}, {8, 3}, {10, 9}},
		{{2, 18}, {5, 8}, {8, 9}, {9, 9}},
	}
	graph, synchronizer := lib.BuildSynchronizedWeightedGraphFromAdjacencyList(adjacencyList, lib.GetGenerator())
	sync_ghs.RunSynchronizedGHS(graph, synchronizer)
}

func TestRunSynchronizedGHSRandom(t *testing.T) {
	checkLogOutput()
	sync_ghs.RunSynchronizedGHSRandom(25, 100, 25)
}

func BenchmarkSynchronizedGHSRandom(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_ghs.RunSynchronizedGHSRandom(25, 100, 25)
	}
}
