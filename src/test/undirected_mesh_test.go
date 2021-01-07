package test

import (
	"io/ioutil"
	"leader/undirected_mesh"
	"lib"
	"log"
	"testing"
)

func TestUndirectedMeshLeader(t *testing.T) {
	checkLogOutput()
	g, n := lib.BuildSynchronizedUndirectedMesh(6, 9)
	undirected_mesh.RunPeterson(g, n, 26)
}

func BenchmarkUndirectedMeshLeader(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		g, n := lib.BuildSynchronizedUndirectedMesh(6, 9)
		undirected_mesh.RunPeterson(g, n, 26)
	}
}
