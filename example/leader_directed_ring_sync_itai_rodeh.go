package main

import (
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/sync_itai_rodeh"
	"os"
	"strconv"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	sync_itai_rodeh.Run(n)
}
