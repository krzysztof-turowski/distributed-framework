package main

import (
	"leader/directed_ring"
	"os"
	"strconv"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	alg, _ := strconv.Atoi(os.Args[len(os.Args)-2])
	if alg == 0 {
		directed_ring.RunDovelKlaweRodeh(n)
	} else if alg == 1 {
		directed_ring.RunDovelKlaweRodehB(n)
	}
}
