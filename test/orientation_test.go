package test

import (
	"io"
	"log"
	"testing"

	"github.com/krzysztof-turowski/distributed-framework/orientation/async_syrotiuk_pachl"
	"github.com/krzysztof-turowski/distributed-framework/orientation/async_syrotiuk_pachl_2"
	"github.com/krzysztof-turowski/distributed-framework/orientation/sync_torus"
)

func TestOrientationSyrotiukPachl(t *testing.T) {
	checkLogOutput()
	async_syrotiuk_pachl.Run(1000)
}

func BenchmarkOrientationSyrotiukPachl(b *testing.B) {
	log.SetOutput(io.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		async_syrotiuk_pachl.Run(1000)
	}
}

func TestOrientationSyrotiukPachl2(t *testing.T) {
	checkLogOutput()
	async_syrotiuk_pachl_2.Run(1000)
}

func BenchmarkOrientationSyrotiukPachl2(b *testing.B) {
	log.SetOutput(io.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		async_syrotiuk_pachl_2.Run(1000)
	}
}

func TestOrientationTorusMans(t *testing.T) {
	checkLogOutput()
	sync_torus.Run(400)
}

func BenchmarkOrientationTorusMans(b *testing.B) {
	log.SetOutput(io.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_torus.Run(10000)
	}
}
