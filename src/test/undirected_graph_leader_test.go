package test

import (
	"leader/undirected_graph"
	"testing"
)

func TestYoYoRandom(t *testing.T) {
	checkLogOutput()
	undirected_graph.RunYoYoRandom(1000, 0.25)
}
