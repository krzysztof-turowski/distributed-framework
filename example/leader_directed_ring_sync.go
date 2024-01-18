package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/sync_chang_roberts"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/sync_dolev_klawe_rodeh"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/sync_itai_rodeh"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/sync_peterson"
)

func main() {
	n, err := strconv.Atoi(os.Args[1])
	if err != nil || n < 1 {
		fmt.Println("Invalid size", n)
		return
	}

	switch algorithm := strings.ToLower(os.Args[2]); algorithm {
	case "chang_roberts":
		sync_chang_roberts.Run(n)
	case "dolev_klawe_rodeh_a":
		sync_dolev_klawe_rodeh.RunA(n)
	case "dolev_klawe_rodeh_b":
		sync_dolev_klawe_rodeh.RunB(n)
	case "itai_rodeh":
		sync_itai_rodeh.Run(n)
	case "peterson":
		sync_peterson.Run(n)
	}
}
