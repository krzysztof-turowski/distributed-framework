package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_graph/sync_casteigts_metivier_robson_zemmari"
)

const exitFailure = 1

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Please specify graph size")
		os.Exit(exitFailure)
	}
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Please specify p")
		os.Exit(exitFailure)
	}

	n, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid size specification")
		os.Exit(exitFailure)
	}
	p, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid p specification")
	}

	sync_casteigts_metivier_robson_zemmari.Run(n, p)
}
