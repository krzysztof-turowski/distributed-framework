package test

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/krzysztof-turowski/distributed-framework/consensus"
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

func TestAgreementPhaseKing(t *testing.T) {
	checkLogOutput()
	v, s := lib.BuildCompleteGraphWithLoops(10, true, lib.GetGenerator())
	consensus.RunPhaseKing(v, s, 3, []int{0, 1, 1, 0, 1, 0, 1, 0, 1, 1})
}

func TestValidityPhaseKing(t *testing.T) {
	checkLogOutput()
	v, s := lib.BuildCompleteGraphWithLoops(10, true, lib.GetGenerator())
	consensus.RunPhaseKing(v, s, 3, []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}

func BenchmarkPhaseKing(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	for iteration := 0; iteration < b.N; iteration++ {
		v, s := lib.BuildCompleteGraphWithLoops(10, true, lib.GetGenerator())
		consensus.RunPhaseKing(v, s, 3, []int{0, 1, 1, 0, 1, 0, 1, 0, 1, 1})
	}
}
