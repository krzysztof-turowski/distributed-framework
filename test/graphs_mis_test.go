package test

import (
	"io"
	"log"
	"testing"

	"github.com/krzysztof-turowski/distributed-framework/graphs/mis/sync_karp_widgerson"
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

func TestKarpWidgerson(t *testing.T) {
	checkLogOutput()
	sync_karp_widgerson.Run(100, 0.75)
}

func BenchmarkKarpWidgerson(b *testing.B) {
	log.SetOutput(io.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_karp_widgerson.Run(100, 0.75)
	}
}
