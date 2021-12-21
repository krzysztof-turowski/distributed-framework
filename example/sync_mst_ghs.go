package main

import (
	"github.com/krzysztof-turowski/distributed-framework/graphs/mst"
	"os"
	"strconv"
)

func main() {
	n, _ := strconv.Atoi(os.Args[1])
	m, _ := strconv.Atoi(os.Args[2])
	maxWeight, _ := strconv.Atoi(os.Args[3])
	mst.RunSynchronizedGHSRandom(n, m, maxWeight)
}
