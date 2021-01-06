package main

import (
	"leader/directed_ring"
	"os"
	"strconv"
	"strings"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	alg := os.Args[len(os.Args)-2]

	alg = strings.ToLower(alg)

	if alg == "a" {
		directed_ring.RunDovelKlaweRodehA(n)
	} else if alg == "b" {
		directed_ring.RunDovelKlaweRodehB(n)
	}
}
