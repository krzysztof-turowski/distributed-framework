package main

import (
	"fmt"
	"github.com/krzysztof-turowski/distributed-framework/leader/clique/async_loui_matsushita_west"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No size specified")
		return
	}
	n, err := strconv.Atoi(os.Args[1])
	if err != nil || n < 2 {
		fmt.Println("Invalid size", n)
		return
	}
	nodes, runner := lib.BuildHamiltonianOrientedCompleteGraph(n)
	async_loui_matsushita_west.Run(nodes, runner)
}
