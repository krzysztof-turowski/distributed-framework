package test

import (
	"github.com/krzysztof-turowski/distributed-framework/leader/ring"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"io/ioutil"
	"log"
	"math/rand"
	"testing"
)

func TestStagesWithFeedback(t *testing.T) {
	checkLogOutput()

	for n := 2; n <= 100; n++ {
		nodes, runner := lib.BuildRing(n)
		ring.RunStagesWithFeedback(nodes, runner)
	}
}

func BenchmarkStagesWithFeedback(b *testing.B) {
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		rand.Seed(0)
		nodes, runner := lib.BuildRing(100)
		ring.RunStagesWithFeedback(nodes, runner)
	}
}
