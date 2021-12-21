package test

import (
	"io/ioutil"
	"leader/directed_hypercube"
	"lib"
	"log"
	"testing"
)

func TestDirectedHypercubeLeader(t *testing.T) {
	checkLogOutput()
	v, s := lib.BuildSynchronizedHypercube(6, true)
	directed_hypercube.RunHyperelect(v, s)
}

func BenchmarkDirectedHypercubeLeader(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		v, s := lib.BuildSynchronizedHypercube(6, true)
		directed_hypercube.RunHyperelect(v, s)
	}
}
