package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/size_estimation/directed_ring/async_itai_rodeh"
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

	r := 1
	if len(os.Args) >= 3 {
		r, err = strconv.Atoi(os.Args[2])
		if err != nil || r < 1 {
			fmt.Println("Invalid confidence requirement:", r)
			return
		}
	}

	messages, time := async_itai_rodeh.Run(n, r)
	fmt.Println("messages:", messages)
	fmt.Println("time:", time)
}
