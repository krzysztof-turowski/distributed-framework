package main

import (
	"fmt"
	"github.com/krzysztof-turowski/distributed-framework/leader/clique/async_humblet"
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No size specified")
		return
	}

	n, err := strconv.Atoi(os.Args[1])
	if err != nil || n < 2 {
		fmt.Println("Invalid size")
		return
	}

	fmt.Println("Building graph...\n")

	nodes, runner := lib.BuildCompleteGraph(n)

	fmt.Println("\nRunning...\n")

	async_humblet.Run(nodes, runner)
}
