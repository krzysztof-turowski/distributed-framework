package main

import (
	"github.com/krzysztof-turowski/distributed-framework/consensus/sync_ben_or"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"os"
	"strconv"
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

	sync_ben_or.Run(vertices, synchronizer, t, V, sync_ben_or.EachMessageRandom(n, t))
}
