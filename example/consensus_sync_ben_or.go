package main

import (
	"github.com/krzysztof-turowski/distributed-framework/consensus/sync_ben_or"
	"os"
	"strconv"
)

func main() {
	N, _ := strconv.Atoi(os.Args[1])
	t, _ := strconv.Atoi(os.Args[2])
	if 5*t >= N {
		panic("t too big, required N > 5 * t")
	}
	processes := make([]byte, N-t)
	behaviours := make([]func(r int) byte, t)

	for i := 0; i < N-t; i++ {
		processes[i] = byte(i % 2)
	}
	for i := 0; i < t; i++ {
		behaviours[i] = func(r int) byte { return 1 }
	}

	sync_ben_or.Run(processes, behaviours)
}
