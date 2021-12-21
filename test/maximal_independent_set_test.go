package test

import (
	"github.com/krzysztof-turowski/distributed-framework/graphs/mis"
	"io/ioutil"
	"log"
	"testing"
)

func TestLuby(t *testing.T) {
	checkLogOutput()
	mis.RunLuby(1000, 0.75)
}

func BenchmarkLuby(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		mis.RunLuby(1000, 0.75)
	}
}
