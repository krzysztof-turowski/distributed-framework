package mst

import (
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"reflect"
	"sort"
	"testing"
)

func compareTreeEdges(expectedTree [][]int, vertices []lib.WeightedGraphNode, t *testing.T) {
	for i, v := range vertices {
		actualTree := getTreeEdges(v)
		sort.Ints(actualTree)
		if !reflect.DeepEqual(expectedTree[i], actualTree) {
			t.Fatalf("For node %d expectedTree tree edges %v but got %v\n", i, expectedTree[i], actualTree)
		}
	}
}

func TestRunSynchronizedGHSTriangle(t *testing.T) {
	adjacencyList := [][][2]int{
		{{2, 10}, {3, 10}},
		{{1, 10}, {3, 20}},
		{{1, 10}, {2, 20}},
	}
	expectedTree := [][]int{
		{2, 3},
		{1},
		{1},
	}
	graph, synchronizer := lib.BuildSynchronizedWeightedGraphFromAdjacencyList(adjacencyList, lib.GetGenerator())
	RunSynchronizedGHS(graph, synchronizer)
	compareTreeEdges(expectedTree, graph, t)
}

func TestRunSynchronizedGHSSimplePath4(t *testing.T) {
	adjacencyList := [][][2]int{
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
	graph, synchronizer := lib.BuildSynchronizedWeightedGraphFromAdjacencyList(adjacencyList, lib.GetGenerator())
	RunSynchronizedGHS(graph, synchronizer)
	compareTreeEdges(expectedTree, graph, t)
}

func TestRunSynchronizedGHSSimplePath8(t *testing.T) {
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
	graph, synchronizer := lib.BuildSynchronizedWeightedGraphFromAdjacencyList(adjacencyList, lib.GetGenerator())
	RunSynchronizedGHS(graph, synchronizer)
	compareTreeEdges(expectedTree, graph, t)
}

func TestRunSynchronizedGHSWikipediaExample(t *testing.T) {
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
	graph, synchronizer := lib.BuildSynchronizedWeightedGraphFromAdjacencyList(adjacencyList, lib.GetGenerator())
	RunSynchronizedGHS(graph, synchronizer)
	compareTreeEdges(expectedTree, graph, t)
}

func TestRunSynchronizedGHSRandom(t *testing.T) {
	RunSynchronizedGHSRandom(25, 100, 25)
}
