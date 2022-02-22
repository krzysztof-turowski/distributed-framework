package test

import (
	"github.com/krzysztof-turowski/distributed-framework/graphs/mis/sync_luby"
	"io/ioutil"
	"log"
	"testing"
)

func TestLuby(t *testing.T) {
	checkLogOutput()
	sync_luby.Run(1000, 0.75)
}

func BenchmarkLuby(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_luby.Run(1000, 0.75)
	}
}
