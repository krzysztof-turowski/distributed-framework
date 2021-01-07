package main

import (
	"leader/undirected_mesh"
	"lib"
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
	undirected_mesh.RunMeshLeader(g, n, (a+b-2)*2)

}
