package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/byzantine/sync_chor_coan"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("To few arguments, expected: [n] [t] where n is the number of processors and t is the maximum number of byzantine processors")
		return
	}

	n, err := strconv.Atoi(os.Args[1])
	if err != nil || n < 1 {
		fmt.Println("Invalid size:", n)
		return
	}

	t, err := strconv.Atoi(os.Args[2])
	if err != nil || 3*t+1 > n || t < 0 {
		fmt.Println("Invalid maximum number of byzantine processors, expected 0 <= t, 3t+1 <= n =", n, "get:", t)
		return
	}

	fmt.Println("Input values (0/1) in format [v0] [v1] ... [v_n-1] with -1 being a byzantine node")

	inputs := make([]int, n)
	byzant := 0
	for i := 0; i < n; i++ {
		fmt.Scanf("%d", &(inputs[i]))
		if inputs[i] != -1 && inputs[i] != 0 && inputs[i] != 1 {
			fmt.Println("Only -1,0,1 accepted as input values, found:", inputs[i])
			panic("Invalid input values")
		}
		if inputs[i] == -1 {
			byzant++
		}
	}

	if byzant > t {
		panic("Too many byzantine nodes")
	}

	nodes, synchronizer := lib.BuildCompleteGraphWithLoops(n, true, lib.GetGenerator())
	messages, time := sync_chor_coan.Run(nodes, synchronizer, t, inputs)
	fmt.Println("messages:", messages)
	fmt.Println("time:", time)
}
