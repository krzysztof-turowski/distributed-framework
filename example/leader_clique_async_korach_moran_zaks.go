package main

import (
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/leader/clique/async_korach_moran_zaks"
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

func upperBound(n int) int {
	return 5*n*int(math.Log2(float64(n))) + 2*n
}

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
	messages, time := async_korach_moran_zaks.Run(nodes, runner)
	fmt.Println("upper bound", upperBound(n))
	fmt.Println("messages:", messages)
	fmt.Println("time:", time)
}
