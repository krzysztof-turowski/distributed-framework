package main

import (
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_mesh"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
	"os"
	"strconv"
)

func main() {
	a, _ := strconv.Atoi(os.Args[len(os.Args)-2])
	b, _ := strconv.Atoi(os.Args[len(os.Args)-1])
	if a <= 2 || b <= 2 {
		log.Println("too small")
		return
	}

	g, n := lib.BuildSynchronizedUndirectedMesh(a, b)
	undirected_mesh.RunPeterson(g, n, (a+b-2)*2)

}
