package test

import (
	"github.com/krzysztof-turowski/distributed-framework/graphs/mis/sync_metivier_a"
	"io"
	"log"
	"testing"

	"github.com/krzysztof-turowski/distributed-framework/graphs/mis/sync_luby"
	"github.com/krzysztof-turowski/distributed-framework/graphs/mis/sync_metivier_c"
)

func TestLuby(t *testing.T) {
	checkLogOutput()
	sync_luby.Run(100, 0.75)
}

func BenchmarkLuby(b *testing.B) {
	log.SetOutput(io.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_luby.Run(100, 0.75)
	}
}

func TestMetivierC(t *testing.T) {
	checkLogOutput()
	sync_metivier_c.Run(100, 0.75)
}

func BenchmarkMetivierC(b *testing.B) {
	log.SetOutput(io.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_metivier_c.Run(100, 0.75)
	}
}

func TestMetivierA(t *testing.T) {
	checkLogOutput()
	sync_metivier_a.Run(100, 0.75)
}

func BenchmarkMetivierA(b *testing.B) {
	log.SetOutput(io.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_metivier_a.Run(100, 0.75)
	}
}
