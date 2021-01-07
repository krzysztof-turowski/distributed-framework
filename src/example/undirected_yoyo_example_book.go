package main

import (
	"leader/undirected_graph"
	"lib"
)

func main() {
	// Graph from Santoro, Design and Analysis, sec. 3.8.3
	adjacencyList := [][]int{
		{7, 9, 10},
		{3, 8, 11},
		{2, 4, 15},
		{3, 16},
		{7, 8, 12},
		{10, 13},
		{1, 5, 9},
		{2, 5, 12},
		{1, 7},
		{1, 6, 13},
		{2, 15, 16},
		{5, 8, 14, 17},
		{6, 10},
		{12, 18, 19},
		{3, 11, 16},
		{4, 11, 15},
		{12, 18, 19},
		{14, 17},
		{14, 17},
	}
	undirected_graph.RunYoYo(lib.BuildSynchronizedGraphFromAdjacencyList(adjacencyList, lib.GetGenerator()))
}
