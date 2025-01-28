package main

import (
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/sync_franklin_2"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	sync_franklin_2.Run(n)
}
