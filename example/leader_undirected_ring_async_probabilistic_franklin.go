package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/async_probabilistic_franklin"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No size specified")
		return
	}
	if len(os.Args) < 3 {
		fmt.Println("No domain size specified")
		return
	}

	n, err := strconv.Atoi(os.Args[1])
	if err != nil || n < 1 {
		fmt.Println("Invalid size", n)
		return
	}

	domainSize, err := strconv.Atoi(os.Args[2])
	if err != nil || domainSize < 2 {
		fmt.Println("Invalid domain size", domainSize)
		return
	}

	async_probabilistic_franklin.Run(n, domainSize)
}
