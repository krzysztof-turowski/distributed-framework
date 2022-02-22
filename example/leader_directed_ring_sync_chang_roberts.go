package main

import (
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/sync_chang_roberts"
	"os"
	"strconv"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	sync_chang_roberts.Run(n)
}
