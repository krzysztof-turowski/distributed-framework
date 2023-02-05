package test

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/krzysztof-turowski/distributed-framework/graphs/ds/sync_lrg"
)

func TestLRG(t *testing.T) {
	checkLogOutput()
	sync_lrg.Run(600, 0.75)
}

func BenchmarkLRG(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_lrg.Run(600, 0.75)
	}
}
