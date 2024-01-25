package test

import (
	"io"
	"log"
	"testing"

	"github.com/krzysztof-turowski/distributed-framework/graphs/mis/sync_luby"
	"github.com/krzysztof-turowski/distributed-framework/graphs/mis/sync_metivier_et_al"
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

func TestMetivierEtAl(t *testing.T) {
	checkLogOutput()
	sync_metivier_et_al.Run(100, 0.75)
}

func BenchmarkMetivierEtAl(b *testing.B) {
	log.SetOutput(io.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_metivier_et_al.Run(100, 0.75)
	}
}
