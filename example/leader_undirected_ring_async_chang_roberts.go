package main

import (
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/async_chang_roberts"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	async_chang_roberts.Run(n)
}
