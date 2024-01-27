package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/orientation/sync_torus"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No size specified")
		return
	}
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	sync_torus.Run(n)
}
