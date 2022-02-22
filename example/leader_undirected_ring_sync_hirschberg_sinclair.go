package main

import (
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/sync_hirschberg_sinclair"
	"os"
	"strconv"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	sync_hirschberg_sinclair.Run(n)
}
