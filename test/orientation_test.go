package test

import (
	"github.com/krzysztof-turowski/distributed-framework/orientation/async_syrotiuk_pachl"
	"io/ioutil"
	"log"
	"testing"
)

func TestOrientationSyrotiukPachl(t *testing.T) {
	checkLogOutput()
	async_syrotiuk_pachl.Run(1000)
}

func BenchmarkOrientationSyrotiukPachl(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		async_syrotiuk_pachl.Run(1000)
	}
}
