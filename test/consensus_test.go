package test

import (
	"github.com/krzysztof-turowski/distributed-framework/consensus/sync_ben_or"
	"github.com/krzysztof-turowski/distributed-framework/consensus/sync_phase_king"
	"github.com/krzysztof-turowski/distributed-framework/consensus/sync_single_bit"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"io/ioutil"
	"log"
	"testing"
)

func TestBenOr(t *testing.T) {
	checkLogOutput()
	f := map[int]int{1: 1, 2: 2}
	V := []int{0, 1, 1, 0, 1, 0, 1, 0, 1, 1, 0}
	for iteration := 0; iteration < 500; iteration++ {
		n, s := lib.BuildCompleteGraphWithLoops(11, true, lib.GetGenerator())
		sync_ben_or.Run(n, s, V, sync_ben_or.GetFaultyBehavior(n, f, &sync_ben_or.Random{}), f)
	}
	for iteration := 0; iteration < 10; iteration++ {
		n, s := lib.BuildCompleteGraphWithLoops(11, true, lib.GetGenerator())
		sync_ben_or.Run(n, s, V, sync_ben_or.GetFaultyBehavior(n, f, &sync_ben_or.Optimal{}), f)
	}
}

func BenchmarkBenOr(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	f := map[int]int{1: 1, 2: 2}
	V := []int{0, 1, 1, 0, 1, 0, 1, 0, 1, 1, 0}
	for iteration := 0; iteration < b.N; iteration++ {
		n, s := lib.BuildCompleteGraphWithLoops(11, true, lib.GetGenerator())
		sync_ben_or.Run(n, s, V, sync_ben_or.GetFaultyBehavior(n, f, &sync_ben_or.Random{}), f)
	}
}

func TestPhaseKing(t *testing.T) {
	checkLogOutput()
	f := map[int]int{1: 1, 2: 2, 3: 3}
	V := []int{0, 1, 1, 0, 1, 0, 1, 0, 1, 1}
	for iteration := 0; iteration < 500; iteration++ {
		n, s := lib.BuildCompleteGraphWithLoops(10, true, lib.GetGenerator())
		sync_phase_king.Run(n, s, V, sync_phase_king.GetFaultyBehavior(n, f, &sync_phase_king.Random{}), f)
	}
	n, s := lib.BuildCompleteGraphWithLoops(10, true, lib.GetGenerator())
	sync_phase_king.Run(n, s, V, sync_phase_king.GetFaultyBehavior(n, f, &sync_phase_king.Optimal{}), f)
}

func BenchmarkPhaseKing(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	f := map[int]int{1: 1, 2: 2}
	V := []int{0, 1, 1, 0, 1, 0, 1, 0, 1, 1, 0}
	for iteration := 0; iteration < b.N; iteration++ {
		n, s := lib.BuildCompleteGraphWithLoops(11, true, lib.GetGenerator())
		sync_phase_king.Run(n, s, V, sync_phase_king.GetFaultyBehavior(n, f, &sync_phase_king.Random{}), f)
	}
}

func TestSingleBit(t *testing.T) {
	checkLogOutput()
	f := map[int]int{1: 1, 2: 2}
	V := []int{0, 1, 1, 0, 1, 0, 1, 0, 1}
	for iteration := 0; iteration < 500; iteration++ {
		n, s := lib.BuildCompleteGraphWithLoops(9, true, lib.GetGenerator())
		sync_single_bit.Run(n, s, V, sync_single_bit.GetFaultyBehavior(n, f, &sync_single_bit.Random{}), f)
	}
	n, s := lib.BuildCompleteGraphWithLoops(9, true, lib.GetGenerator())
	sync_single_bit.Run(n, s, V, sync_single_bit.GetFaultyBehavior(n, f, &sync_single_bit.Optimal{}), f)
}

func BenchmarkSingleBit(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	f := map[int]int{1: 1, 2: 2}
	V := []int{0, 1, 1, 0, 1, 0, 1, 0, 1, 1, 0}
	for iteration := 0; iteration < b.N; iteration++ {
		n, s := lib.BuildCompleteGraphWithLoops(11, true, lib.GetGenerator())
		sync_single_bit.Run(n, s, V, sync_single_bit.GetFaultyBehavior(n, f, &sync_single_bit.Random{}), f)
	}
}
