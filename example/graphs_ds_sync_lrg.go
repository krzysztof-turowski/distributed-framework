package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/graphs/ds/sync_lrg"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-2])
	p, _ := strconv.ParseFloat(os.Args[len(os.Args)-1], 64)
	fmt.Println(n, p)
	sync_lrg.Run(n, p)
}
