package main

import (
	"fmt"
	"os"
	"strconv"
	
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring"
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

	fmt.Println("Building Directed Ring...")

	nodes, runner := lib.BuildDirectedRing(n)

	fmt.Println("\nRunning...")

	directed_ring.RunAsyncItaiRodeh(nodes, runner)

	fmt.Println("\nFinished")
}
