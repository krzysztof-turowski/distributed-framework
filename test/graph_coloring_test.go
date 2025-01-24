package test

import (
	"io"
	"log"
	"testing"
	

	"github.com/krzysztof-turowski/distributed-framework/graphs/coloring/sync_goldberg_plotkin_shannon"
)

func TestGps(t *testing.T) {
	checkLogOutput()
	sync_goldberg_plotkin_shannon.Run(100, 0.75)

}

func BenchmarkGps(b *testing.B) {
	log.SetOutput(io.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_goldberg_plotkin_shannon.Run(100, 0.75)
	}
}
