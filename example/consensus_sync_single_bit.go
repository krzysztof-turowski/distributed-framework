package main

import (
	"github.com/krzysztof-turowski/distributed-framework/consensus/sync_single_bit"
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

func main() {
	n, _ := strconv.Atoi(os.Args[1])
	t, _ := strconv.Atoi(os.Args[2])
	var V []int
	for i := 3; i < n+3; i++ {
		v, _ := strconv.Atoi(os.Args[i])
		V = append(V, v)
	}
	nodes, synchronizer := lib.BuildCompleteGraphWithLoops(n, true, lib.GetGenerator())
	faultyIndices := make(map[int]int)
	for i := n + 3; i < n+3+t; i++ {
		x, _ := strconv.Atoi(os.Args[i])
		faultyIndices[x] = len(faultyIndices) + 1
	}
	var strategy sync_single_bit.Strategy
	switch os.Args[n+3+t] {
	case "Random":
		strategy = &sync_single_bit.Random{}
	case "Optimal":
		strategy = &sync_single_bit.Optimal{}
	default:
		panic("Strategy not supported")
	}

	sync_single_bit.Run(nodes, synchronizer, V,
		sync_single_bit.GetFaultyBehavior(nodes, faultyIndices, strategy), faultyIndices)
}
