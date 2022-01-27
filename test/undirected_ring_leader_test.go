package test

import (
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_ring"
	"io/ioutil"
	"log"
	"testing"
)

func TestUndirectedRingHirschbergSinclair(t *testing.T) {
	checkLogOutput()
	undirected_ring.RunHirschbergSinclair(1000)
}

func BenchmarkUndirectedRingHirschbergSinclair(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		undirected_ring.RunHirschbergSinclair(1000)
	}
}