package test

import (
	"github.com/krzysztof-turowski/distributed-framework/consensus/sync_ben_or"
	"github.com/krzysztof-turowski/distributed-framework/consensus/sync_phase_king"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"io/ioutil"
	"log"
	"testing"
)

func TestAllCorrectBenOr(t *testing.T) {
	checkLogOutput()
	n := 100

	processes := make([]byte, n)
	for i := 0; i < n; i++ {
		processes[i] = byte(i % 2)
	}
	sync_ben_or.Run(processes, make([]func(r int) byte, 0))
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
	sync_ben_or.Run(processes, behaviours)
}

func TestSmallBenOr(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	// N == 6, t == 1
	processes := []byte{0, 1, 0, 1, 0}
	behaviours := []func(r int) byte{func(r int) byte { return 1 }}

	for it := 0; it < 100; it++ {
		sync_ben_or.Run(processes, behaviours)
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
	sync_ben_or.Run(processes, behaviours)
}

func BenchmarkBenOr(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	n, f := 49, 7

	processes := make([]byte, n-f)
	for i := 0; i < n-f; i++ {
		processes[i] = byte(1 - i%2)
	}
	behaviours := make([]func(r int) byte, f)
	for i := 0; i < f; i++ {
		behaviours[i] = func(r int) byte { return byte((i + r) % 2) }
	}

	for it := 0; it < b.N; it++ {
		sync_ben_or.Run(processes, behaviours)
	}
}

func TestAgreementPhaseKing(t *testing.T) {
	checkLogOutput()
	v, s := lib.BuildCompleteGraphWithLoops(10, true, lib.GetGenerator())
	sync_phase_king.RunPhaseKing(v, s, 3, []int{0, 1, 1, 0, 1, 0, 1, 0, 1, 1})
}

func TestValidityPhaseKing(t *testing.T) {
	checkLogOutput()
	v, s := lib.BuildCompleteGraphWithLoops(10, true, lib.GetGenerator())
	sync_phase_king.RunPhaseKing(v, s, 3, []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}

func BenchmarkPhaseKing(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		v, s := lib.BuildCompleteGraphWithLoops(10, true, lib.GetGenerator())
		sync_phase_king.RunPhaseKing(v, s, 3, []int{0, 1, 1, 0, 1, 0, 1, 0, 1, 1})
	}
}
