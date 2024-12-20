package test

import (
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/Skryg/distributed-framework/leader/directed_ring/async_itah_rodeh_2"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/async_higham_przytycka"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/async_itah_rodeh"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/sync_chang_roberts"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/sync_dolev_klawe_rodeh"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/sync_itai_rodeh"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/sync_peterson"
)

func TestDirectedRingChangRoberts(t *testing.T) {
	checkLogOutput()
	sync_chang_roberts.Run(1000)
}

func TestDirectedRingItaiRodeh(t *testing.T) {
	checkLogOutput()
	sync_itai_rodeh.Run(1000)
}

func TestAsyncDirectedRingItaiRodeh2(t *testing.T) {
	checkLogOutput()
	async_itah_rodeh_2.RunDefault(5000)
}

const PROB = 0.01

type RandBit struct {
	rand *rand.Rand
}

func (r RandBit) ReadBit() uint8 {
	if rand.Float64() > PROB {
		return 0
	}
	return 1
}

func getRandBit() async_itah_rodeh_2.IRandomBit {
	return RandBit{
		rand.New(rand.NewSource(time.Now().UnixMilli())),
	}
}
func TestAsyncDirectedRingItaiRodehRand(b *testing.T) {
	checkLogOutput()
	async_itah_rodeh_2.Run(5000, getRandBit())
}

func TestDirectedRingRunDolevKlaweRodehA(t *testing.T) {
	checkLogOutput()
	sync_dolev_klawe_rodeh.RunA(1000)
}

func TestDirectedRingRunDolevKlaweRodehB(t *testing.T) {
	checkLogOutput()
	sync_dolev_klawe_rodeh.RunB(1000)
}

func TestDirectedRingPeterson(t *testing.T) {
	checkLogOutput()
	sync_peterson.Run(1000)
}

func TestDirectedRingHighamPrzytycka(b *testing.T) {
	checkLogOutput()
	async_higham_przytycka.Run(1000)
}

func BenchmarkDirectedRingChangRoberts(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_chang_roberts.Run(1000)
	}
}

func BenchmarkDirectedRingItaiRodeh(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_itai_rodeh.Run(1000)
	}
}

func BenchmarkDirectedRingRunDolevKlaweRodehA(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_dolev_klawe_rodeh.RunA(1000)
	}
}

func BenchmarkDirectedRingRunDolevKlaweRodehB(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_dolev_klawe_rodeh.RunB(1000)
	}
}

func BenchmarkDirectedRingPeterson(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		sync_peterson.Run(1000)
	}
}

func BenchmarkAsyncDirectedRingItaiRodeh(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		async_itah_rodeh.Run(1000)
	}
}

func BenchmarkAsyncDirectedRingItaiRodeh2(b *testing.B) {
	log.SetOutput(io.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		async_itah_rodeh_2.RunDefault(1000)
	}
}

func BenchmarkAsyncDirectedRingHighamPrzytycka(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		async_higham_przytycka.Run(1000)
	}
}
