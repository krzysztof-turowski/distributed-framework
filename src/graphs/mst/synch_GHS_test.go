package mst

import (
	"io/ioutil"
	"lib"
	"log"
	"reflect"
	"sort"
	"testing"
)

func compareTreeEdges(expectedTree [][]int, vertices []lib.WeightedGraphNode, t *testing.T) {
	for i, v := range vertices {
		l := getTreeEdges(v)
		sort.Ints(l)
		if !reflect.DeepEqual(expectedTree[i], l) {
			t.Fatalf("For %d expected tree edges %v but got %v\n", i, expectedTree[i], l)
		}
	}
}

func TestRunSynchGHSTriangle(t *testing.T) {
	adjencyList := [][][2]int{
		{{2, 10}, {3, 10}},
		{{1, 10}, {3, 20}},
		{{1, 10}, {2, 20}},
	}
	expectedTree := [][]int{
		{2, 3},
		{1},
		{1},
	}
	graph, synchronizer := lib.BuildSynchronizedWeightedGraphFromAdjencyList(adjencyList, lib.GetGenerator())
	RunSynchGHS(graph, synchronizer)
	compareTreeEdges(expectedTree, graph, t)
}

func TestRunSynchGHSSimplePath4(t *testing.T) {
	adjencyList := [][][2]int{
		{{2, 10}},
		{{1, 10}, {3, 20}},
		{{2, 20}, {4, 10}},
		{{3, 10}},
	}
	expectedTree := [][]int{
		{2},
		{1, 3},
		{2, 4},
		{3},
	}
	graph, synchronizer := lib.BuildSynchronizedWeightedGraphFromAdjencyList(adjencyList, lib.GetGenerator())
	RunSynchGHS(graph, synchronizer)
	compareTreeEdges(expectedTree, graph, t)
}

func TestRunSynchGHSSimplePath8(t *testing.T) {
	adjencyList := [][][2]int{
		{{2, 10}},
		{{1, 10}, {3, 20}},
		{{2, 20}, {4, 10}},
		{{3, 10}, {5, 30}},
		{{4, 30}, {6, 10}},
		{{5, 10}, {7, 20}},
		{{6, 20}, {8, 10}},
		{{7, 10}},
	}
	expectedTree := [][]int{
		{2},
		{1, 3},
		{2, 4},
		{3, 5},
		{4, 6},
		{5, 7},
		{6, 8},
		{7},
	}
	graph, synchronizer := lib.BuildSynchronizedWeightedGraphFromAdjencyList(adjencyList, lib.GetGenerator())
	RunSynchGHS(graph, synchronizer)
	compareTreeEdges(expectedTree, graph, t)
}

func TestRunSynchGHSWikipediaExample(t *testing.T) {
	adjencyList := [][][2]int{
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
	expectedTree := [][]int{
		{3, 4},
		{3},
		{1, 2},
		{1, 5},
		{4, 7, 10},
		{7},
		{5, 6, 8},
		{7, 9},
		{8},
		{5},
	}
	graph, synchronizer := lib.BuildSynchronizedWeightedGraphFromAdjencyList(adjencyList, lib.GetGenerator())
	RunSynchGHS(graph, synchronizer)
	compareTreeEdges(expectedTree, graph, t)
}

func TestRunSynchGHSRandom(t *testing.T) {
	RunSynchGHSRandom(25, 100, 25)
}

func BenchmarkSynchGHSRandom(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		RunSynchGHSRandom(25, 100, 25)
	}
}
