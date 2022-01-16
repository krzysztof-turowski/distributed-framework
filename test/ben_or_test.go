package test

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/krzysztof-turowski/distributed-framework/consensus/sync_ben_or"
)

func TestAllCorrectBenOr(t *testing.T) {
	checkLogOutput()
	n := 100

	processes := make([]byte, n)
	for i := 0; i < n; i++ {
		processes[i] = byte(i % 2)
	}
	ben_or.RunBenOr(processes, make([]func(r int) byte, 0))
}

func TestSameStartingBenOr(t *testing.T) {
	checkLogOutput()
	n, f := 100, 9

	processes := make([]byte, n-f)
	for i := 0; i < n-f; i++ {
		processes[i] = 0
	}
	behaviours := make([]func(r int) byte, f)
	for i := 0; i < f; i++ {
		behaviours[i] = func(r int) byte { return 1 }
	}
	ben_or.RunBenOr(processes, behaviours)
}

func TestSmallBenOr(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	// N == 6, t == 1
	processes := []byte{0, 1, 0, 1, 0}
	behaviours := []func(r int) byte{func(r int) byte { return 1 }}

	for it := 0; it < 100; it++ {
		ben_or.RunBenOr(processes, behaviours)
	}
}

func TestBigBenOr(t *testing.T) {
	checkLogOutput()
	n, f := 101, 20

	processes := make([]byte, n-f)
	for i := 0; i < n-f; i++ {
		processes[i] = byte(i % 2)
	}
	behaviours := make([]func(r int) byte, f)
	for i := 0; i < f; i++ {
		behaviours[i] = func(r int) byte { return byte((i + r) % 2) }
	}
	ben_or.RunBenOr(processes, behaviours)
}

func BenchmarkBenOr(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	n, f := 49, 7

	processes := make([]byte, n-f)
	for i := 0; i < n-f; i++ {
		processes[i] = byte(1 - i % 2)
	}
	behaviours := make([]func(r int) byte, f)
	for i := 0; i < f; i++ {
		behaviours[i] = func(r int) byte { return byte((i + r) % 2) }
	}

	for it := 0; it < b.N; it++ {
		ben_or.RunBenOr(processes, behaviours)
	}
}
