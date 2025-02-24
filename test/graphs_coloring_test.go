package test

import (
	"io"
	"log"
	"testing"
	"github.com/krzysztof-turowski/distributed-framework/graphs/coloring/sync_gps"
)

func TestGps(t *testing.T) {
	checkLogOutput()
	sync_gps.Run(100, 0.75)
}

func BenchmarkGps(b *testing.B) {
	log.SetOutput(io.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_gps.Run(100, 0.75)
	}
}
