package main

import (
	"fmt"
	"os"
	"strconv"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"github.com/krzysztof-turowski/distributed-framework/graphs/ds/sync_kuhn_wattenhofer"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("use random graph size(number of vertices) int, edge density float64, roundsParameter int")
		return
	}
	n, errn := strconv.Atoi(os.Args[1])
	p, errp := strconv.ParseFloat(os.Args[2], 64)
	k, errk := strconv.Atoi(os.Args[3])
	if errn != nil || errp != nil || errk != nil {
		fmt.Println("bad parameters types")
		fmt.Println("use random graph size(number of vertices) int, edge density float64, roundsParameter int")
		return
	}
	vertices, synchronizer := lib.BuildSynchronizedRandomGraph(n, p)
	sync_kuhn_wattenhofer.Run(vertices, synchronizer, k)
}
