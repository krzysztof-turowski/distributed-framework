package test

import (
	"graphs"
	"io/ioutil"
	"log"
	"testing"
)

func TestLubyMIS(t *testing.T) {
	checkLogOutput()
	graphs.RunLubyMIS(1000, 0.75)
}

func BenchmarkLubyMIS(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		graphs.RunLubyMIS(1000, 0.75)
	}
}
