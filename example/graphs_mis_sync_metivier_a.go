package main

import (
	"fmt"
	"github.com/krzysztof-turowski/distributed-framework/graphs/mis/sync_metivier_a"
	"os"
	"strconv"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-2])
	p, _ := strconv.ParseFloat(os.Args[len(os.Args)-1], 64)
	fmt.Println(n, p)
	sync_metivier_a.Run(n, p)
}
