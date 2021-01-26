package main

import (
	"graphs/mst"
	"os"
	"strconv"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-2])
	m, _ := strconv.Atoi(os.Args[len(os.Args)-2])
	maxWeight, _ := strconv.Atoi(os.Args[len(os.Args)-2])
	mst.RunSynchGHSRandom(n, m, maxWeight)
}
