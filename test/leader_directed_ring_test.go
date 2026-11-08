package test

import (
	"io"
	"io/ioutil"
	"log"
	"testing"

	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/async_higham_przytycka"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/async_itah_rodeh"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/async_itai_rodeh_2"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/sync_chang_roberts"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/sync_dolev_klawe_rodeh"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/sync_itai_rodeh"
	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/sync_itai_rodeh_improved"
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

func TestDirectedRingItaiRodehImproved(t *testing.T) {
	checkLogOutput()
	sync_itai_rodeh_improved.Run(1000)
	for _, q := range []float64{0.5, 1, 2} {
		sync_itai_rodeh_improved.RunWithParameter(100, q)
	}
}

// Fails with little probability
func TestAsyncDirectedRingItaiRodeh2(t *testing.T) {
	checkLogOutput()
	async_itai_rodeh_2.Run(5000)
}

func TestAsyncDirectedRingItaiRodehRand(b *testing.T) {
	checkLogOutput()
	n := 2000
	async_itai_rodeh_2.RunProb(n, 0.1)
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

func BenchmarkDirectedRingItaiRodehImproved(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	n, messages, rounds := 64, 0, 0
	for iteration := 0; iteration < b.N; iteration++ {
		m, r := sync_itai_rodeh_improved.RunWithParameter(n, 1)
		messages, rounds = messages+m, rounds+r
	}

	b.ReportMetric(float64(rounds)/float64(b.N*n), "phases/op")
	b.ReportMetric(float64(messages)/float64(b.N*n), "messages/node/op")
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
		async_itai_rodeh_2.Run(1000)
	}
}

func BenchmarkAsyncDirectedRingHighamPrzytycka(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		async_higham_przytycka.Run(1000)
	}
}
