package main

import (
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/sync_franklin"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	sync_franklin.Run(n)
}
