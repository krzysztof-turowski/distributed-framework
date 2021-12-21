package test

import (
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_graph"
	"io/ioutil"
	"log"
	"testing"
)

func TestYoYoRandom(t *testing.T) {
	checkLogOutput()
	undirected_graph.RunYoYoRandom(1000, 0.25)
}

func BenchmarkYoYoRandom(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		undirected_graph.RunYoYoRandom(1000, 0.25)
	}
}
