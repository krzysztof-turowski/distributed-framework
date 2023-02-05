package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/async_franklin"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No size specified")
		return
	}

	n, err := strconv.Atoi(os.Args[1])
	if err != nil || n < 1 {
		fmt.Println("Invalid size", n)
		return
	}

	fmt.Println("Building graph...")

	nodes, runner := lib.BuildRing(n)

	fmt.Println("\nRunning...")

	async_franklin.Run(nodes, runner)

	fmt.Println("\nFinished")
}
