package main

import (
	"github.com/krzysztof-turowski/distributed-framework/size_estimation/directed_ring/async_itai_rodeh_2"
	"os"
	"strconv"
)

func main() {
	var num, confidence int = 0, 1
	if len(os.Args) < 2 || len(os.Args) > 3 {
		panic("Invalid number of arguments")
	}
	res, err := strconv.Atoi(os.Args[1])
	if err != nil || res < 2 {
		panic("Invalid number of processors")
	}
	num = res
	if len(os.Args) == 3 {
		res, err = strconv.Atoi(os.Args[2])
		if err != nil || res < 1 {
			panic("Invalid confidence")
		}
		confidence = res
	} else {
		confidence = 1
	}
	async_itai_rodeh_2.Run(num, confidence)
}
