package main

import (
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring"
	"os"
	"strconv"
)

func main(){
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	undirected_ring.RunHirschbergSinclair(n)
}
