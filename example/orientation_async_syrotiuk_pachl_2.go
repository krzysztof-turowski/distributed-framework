package main

import (
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/orientation/async_syrotiuk_pachl_2"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	async_syrotiuk_pachl_2.Run(n)
}
