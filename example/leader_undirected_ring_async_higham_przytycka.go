package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/HubertBer/distributed-framework/leader/undirected_ring/async_higham_przytycka"
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Number of nodes not provided!")
		return
	}

	n, err := strconv.Atoi(os.Args[1])

	if err != nil {
		fmt.Println("Invalid number of nodes!")
		return
	}

	fmt.Println("Building ring...\n")

	nodes, runner := lib.BuildRing(n)

	fmt.Println("Running...\n")

	async_higham_przytycka.Run(nodes, runner)

	fmt.Println("Finished")
}
