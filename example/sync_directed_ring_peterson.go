package main

import (
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	directed_ring.RunPeterson(n)
}
