package test

import (
	"io"
	"log"
	"os"
	"testing"

	"github.com/krzysztof-turowski/distributed-framework/orientation/async_syrotiuk_pachl"
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

func TestOrientationTorusMans(t *testing.T) {
	// checkLogOutput()
	f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	sync_torus.Run(25)
}

func BenchmarkOrientationTorusMans(b *testing.B) {
	log.SetOutput(io.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_torus.Run(100)
	}
}
