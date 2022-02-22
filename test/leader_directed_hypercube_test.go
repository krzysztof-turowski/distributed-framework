package test

import (
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_hypercube/sync_hyperelect"
	"io/ioutil"
	"log"
	"testing"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

func TestDirectedHypercubeLeader(t *testing.T) {
	checkLogOutput()
	v, s := lib.BuildSynchronizedHypercube(6, true)
	sync_hyperelect.Run(v, s)
}

func BenchmarkDirectedHypercubeLeader(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		v, s := lib.BuildSynchronizedHypercube(6, true)
		sync_hyperelect.Run(v, s)
	}
}
