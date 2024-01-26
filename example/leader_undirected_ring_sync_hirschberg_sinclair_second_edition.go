package main

import (
	pathToCandidate "github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring/sync_hirschberg_sinclair_second_edition"
	"log"
	"os"
	"strconv"
)

func main() {
	N, _ := strconv.Atoi(os.Args[1])
	if !checkN(N) {
		log.Println("Value of N must be more than 0.")
		return
	}

	argsArr := pathToCandidate.GenerateShuffledArray(N)
	pathToCandidate.BuildUndirectedRingAndRun(N, argsArr)
}

func checkN(N int) bool {
	if N < 1 {
		return false
	}
	return true
}

func checkArgsArr(argsArr []int) bool {
	uniqueVals := make(map[int]bool)

	for _, num := range argsArr {
		if uniqueVals[num] {
			return false
		}
		uniqueVals[num] = true
	}

	return true
}
