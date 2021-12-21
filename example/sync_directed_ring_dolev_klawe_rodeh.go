package main

import (
	"distributed-framework/leader/directed_ring"
	"os"
	"strconv"
	"strings"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	switch algorithm := strings.ToLower(os.Args[len(os.Args)-2]); algorithm {
	case "a":
		directed_ring.RunDolevKlaweRodehA(n)
	case "b":
		directed_ring.RunDolevKlaweRodehB(n)
	}
}
