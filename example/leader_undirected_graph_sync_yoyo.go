package main

import (
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_graph/sync_yoyo"
	"os"
	"strconv"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-2])
	p, _ := strconv.ParseFloat(os.Args[len(os.Args)-1], 64)
	sync_yoyo.RunRandom(n, p)
}
