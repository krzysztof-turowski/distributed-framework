package main

import (
	"distributed-framework/leader/directed_ring"
	"os"
	"strconv"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	directed_ring.RunItaiRodeh(n)
}
