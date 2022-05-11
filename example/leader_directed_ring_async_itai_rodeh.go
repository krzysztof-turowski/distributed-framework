package main

import (
	"fmt"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/async_itah_rodeh"
	"os"
	"strconv"
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

	async_itah_rodeh.Run(n)
}
