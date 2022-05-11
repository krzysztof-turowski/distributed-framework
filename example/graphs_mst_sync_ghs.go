package main

import (
	"github.com/krzysztof-turowski/distributed-framework/graphs/mst/sync_ghs"
	"os"
	"strconv"
)

func main() {
	n, _ := strconv.Atoi(os.Args[1])
	m, _ := strconv.Atoi(os.Args[2])
	maxWeight, _ := strconv.Atoi(os.Args[3])
	sync_ghs.RunSynchronizedGHSRandom(n, m, maxWeight)
}
