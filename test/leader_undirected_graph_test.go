package test

import (
	"io"
	"log"
	"testing"

	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_graph/sync_casteigts_metivier_robson_zemmari"
	"github.com/krzysztof-turowski/distributed-framework/leader/undirected_graph/sync_yoyo"
)

func TestYoYoRandom(t *testing.T) {
	checkLogOutput()
	sync_yoyo.RunRandom(100, 0.25)
}

func BenchmarkYoYoRandom(b *testing.B) {
	log.SetOutput(io.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_yoyo.RunRandom(100, 0.25)
	}
}

func TestCasteigtsEtAl(t *testing.T) {
	checkLogOutput()
	sync_casteigts_metivier_robson_zemmari.Run(100, 0.25)
}

func BenchmarkCasteigtsEtAl(b *testing.B) {
	log.SetOutput(io.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_casteigts_metivier_robson_zemmari.Run(100, 0.25)
	}
}
