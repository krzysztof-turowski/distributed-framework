package main

import (
	"graphs/mis"
	"os"
	"strconv"
)

func main() {
	n, _ := strconv.Atoi(os.Args[len(os.Args)-2])
	p, _ := strconv.ParseFloat(os.Args[len(os.Args)-1], 64)
	mis.RunLuby(n, p)
}
