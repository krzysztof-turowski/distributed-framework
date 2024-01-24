package test

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/krzysztof-turowski/distributed-framework/size_estimation/directed_ring/async_itai_rodeh_2"

	"github.com/krzysztof-turowski/distributed-framework/size_estimation/directed_ring/async_itai_rodeh"
)

func TestItaiRodeh(t *testing.T) {
	checkLogOutput()
	async_itai_rodeh.Run(1000, 1)
}

func BenchmarkItaiRodeh(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		async_itai_rodeh.Run(1000, 1)
	}
}

func TestItaiRodeh2(t *testing.T) {
	checkLogOutput()
	async_itai_rodeh_2.Run(1000, 2)
}

func BenchmarkItaiRodeh2(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		async_itai_rodeh_2.Run(1000, 1)
	}
}
