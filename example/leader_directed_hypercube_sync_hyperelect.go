package main

import (
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_hypercube/sync_hyperelect"
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

func main() {
	dim, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	oriented := true
	sync_hyperelect.Run(lib.BuildSynchronizedHypercube(dim, oriented))
}
