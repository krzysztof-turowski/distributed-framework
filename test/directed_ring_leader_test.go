package test

import (
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring"
	"io/ioutil"
	"log"
	"testing"
)

func TestDirectedRingChangRoberts(t *testing.T) {
	checkLogOutput()
	directed_ring.RunChangRoberts(1000)
}

func TestDirectedRingItaiRodeh(t *testing.T) {
	checkLogOutput()
	directed_ring.RunItaiRodeh(1000)
}

func TestDirectedRingRunDolevKlaweRodehA(t *testing.T) {
	checkLogOutput()
	directed_ring.RunDolevKlaweRodehA(1000)
}

func TestDirectedRingRunDolevKlaweRodehB(t *testing.T) {
	checkLogOutput()
	directed_ring.RunDolevKlaweRodehB(1000)
}

func TestDirectedRingPeterson(t *testing.T) {
	checkLogOutput()
	directed_ring.RunPeterson(1000)
}

func BenchmarkDirectedRingChangRoberts(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		directed_ring.RunChangRoberts(1000)
	}
}

func BenchmarkDirectedRingItaiRodeh(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		directed_ring.RunItaiRodeh(1000)
	}
}

func BenchmarkDirectedRingRunDolevKlaweRodehA(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		directed_ring.RunDolevKlaweRodehA(1000)
	}
}

func BenchmarkDirectedRingRunDolevKlaweRodehB(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		directed_ring.RunDolevKlaweRodehB(1000)
	}
}

func BenchmarkDirectedRingPeterson(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		directed_ring.RunPeterson(1000)
	}
}

func BenchmarkAsyncDirectedRingItaiRodeh(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		directed_ring.RunAsyncItaiRodeh(1000)
	}
}
