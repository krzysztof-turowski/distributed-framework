package main

import (
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/leader/directed_hypercube"
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

func main() {
	dim, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	oriented := true
	directed_hypercube.RunHyperelect(lib.BuildSynchronizedHypercube(dim, oriented))
}
