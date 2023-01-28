package main

import (
	"github.com/krzysztof-turowski/distributed-framework/orientation/async_syrotiuk_pachl"
	"os"
	"strconv"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	async_syrotiuk_pachl.Run(n)
}