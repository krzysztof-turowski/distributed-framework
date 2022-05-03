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
	vertices, synchronizer := lib.BuildCompleteGraphWithLoops(n, true, lib.GetGenerator())

	sync_single_bit.Run(vertices, synchronizer, t, V, sync_single_bit.EachMessageRandom(n, t))
}
