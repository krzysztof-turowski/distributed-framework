package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/leader/clique"
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

	leader := clique.RunHumblet(nodes, runner)

	fmt.Println("\nLeader:", leader)
}
