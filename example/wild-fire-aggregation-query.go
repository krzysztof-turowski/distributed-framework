package main

import (
	"fmt"
	"os"
	"strconv"

	wildfire "github.com/krzysztof-turowski/distributed-framework/aggregation/wild_fire"
)

func main() {
	message := "For instance you can call: wild-fire-aggregation-query 500 0.01 0.001 max 5"

	n := 500                   // 1
	edgeProbability := 0.01    // 2
	faultyProbability := 0.001 // 3
	query := "max"             // 4
	diffValues := 50           // 5
	diameterExtimation := 5    // 6
	var err error

	if len(os.Args) >= 6 {
		diameterExtimation, err = strconv.Atoi(os.Args[6])
		if err != nil || diameterExtimation < 1 {
			fmt.Println("Invalid diameterExtimation:", diameterExtimation)
			fmt.Println(message)
			return
		}
	}
	if len(os.Args) >= 5 {
		diffValues, err = strconv.Atoi(os.Args[5])
		if err != nil || diffValues < 1 {
			fmt.Println("Invalid diameterExtimation:", diffValues)
			fmt.Println(message)
			return
		}
	}
	if len(os.Args) >= 4 {
		query = os.Args[4]
		if query != "max" && query != "min" {
			fmt.Println("Invalid query:", query)
			fmt.Println(message)
			return
		}
	}
	if len(os.Args) >= 3 {
		faultyProbability, err = strconv.ParseFloat(os.Args[3], 32)
		if err != nil || faultyProbability < 0 || faultyProbability > 1 {
			fmt.Println("Invalid faultyProbability:", faultyProbability)
			fmt.Println(message)
			return
		}
	}
	if len(os.Args) >= 2 {
		edgeProbability, err = strconv.ParseFloat(os.Args[2], 32)
		if err != nil || edgeProbability < 0 || edgeProbability > 1 {
			fmt.Println("Invalid faultyProbability:", edgeProbability)
			fmt.Println(message)
			return
		}
	}
	if len(os.Args) >= 1 {
		n, err = strconv.Atoi(os.Args[1])
		if err != nil || n < 1 {
			fmt.Println("Invalid n:", n)
			fmt.Println(message)
			return
		}
	}

	a, b := wildfire.Run(n, edgeProbability, faultyProbability, query, diffValues, diameterExtimation)
	fmt.Println("Query result:", a, ". #(messages): ", b)
}
