package main

import (
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/async_hirschberg_sinclair_2"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	async_hirschberg_sinclair_2.Run(n)
}
