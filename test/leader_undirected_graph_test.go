package test

import (
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_graph/sync_yoyo"
	"io/ioutil"
	"log"
	"testing"
)

func TestYoYoRandom(t *testing.T) {
	checkLogOutput()
	sync_yoyo.RunRandom(1000, 0.25)
}

func BenchmarkYoYoRandom(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_yoyo.RunRandom(1000, 0.25)
	}
}
