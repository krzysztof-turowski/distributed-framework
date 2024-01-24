package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/leader/clique/async_afek_gafni_b"
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No size specified")
		return
	}

	n, err := strconv.Atoi(os.Args[1])
	if err != nil || n < 1 {
		fmt.Println("Invalid size:", n)
		return
	}

	nodes, runner := lib.BuildCompleteGraph(n)
	messages, time := async_afek_gafni_b.Run(nodes, runner)
	fmt.Println("messages:", messages)
	fmt.Println("time:", time)
}
