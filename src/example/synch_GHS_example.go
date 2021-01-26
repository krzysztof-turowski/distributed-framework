package main

import (
	"graphs/mst"
	"lib"
)

func main() {
	weightedAdjencyList := [][][2]int{
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
	mst.RunSynchGHS(lib.BuildSynchronizedWeightedGraphFromAdjencyList(adjencyList, lib.GetGenerator()))
}
