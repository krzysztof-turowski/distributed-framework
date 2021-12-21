package main

import (
	"leader/directed_hypercube"
	"lib"
	"os"
	"strconv"
)

func main() {
	dim, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	oriented := true
	directed_hypercube.RunHyperelect(lib.BuildSynchronizedHypercube(dim, oriented))
}
