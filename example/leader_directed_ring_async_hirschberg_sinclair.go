package main

import (
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/async_hirschberg_sinclair"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	async_hirschberg_sinclair.Run(n)
}
