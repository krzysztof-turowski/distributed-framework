package test

import (
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_mesh/sync_peterson"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"io/ioutil"
	"log"
	"testing"
)

func TestUndirectedMeshLeader(t *testing.T) {
	checkLogOutput()
	g, n := lib.BuildSynchronizedUndirectedMesh(6, 9)
	sync_peterson.Run(g, n, 26)
}

func BenchmarkUndirectedMeshLeader(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		g, n := lib.BuildSynchronizedUndirectedMesh(6, 9)
		sync_peterson.Run(g, n, 26)
	}
}
