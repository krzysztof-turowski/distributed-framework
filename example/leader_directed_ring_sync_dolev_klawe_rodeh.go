package main

import (
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/sync_dolev_klawe_rodeh"
	"os"
	"strconv"
	"strings"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	switch algorithm := strings.ToLower(os.Args[len(os.Args)-2]); algorithm {
	case "a":
		sync_dolev_klawe_rodeh.RunA(n)
	case "b":
		sync_dolev_klawe_rodeh.RunB(n)
	}
}
