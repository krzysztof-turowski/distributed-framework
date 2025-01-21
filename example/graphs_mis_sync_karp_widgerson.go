package main

import (
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/graphs/mis/sync_karp_widgerson"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-2])
	p, _ := strconv.ParseFloat(os.Args[len(os.Args)-1], 64)
	sync_karp_widgerson.Run(n, p)
}
