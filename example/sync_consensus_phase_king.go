package main

import (
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/consensus"
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
	vertices, synchronizer := lib.BuildCompleteGraphWithLoops(n, true, lib.GetGenerator())

	consensus.RunPhaseKing(vertices, synchronizer, t, V)
}
