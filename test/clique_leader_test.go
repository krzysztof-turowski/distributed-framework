package test

import (
	"testing"
	"math/rand"
	"log"
	"io/ioutil"
	"github.com/krzysztof-turowski/distributed-framework/leader/clique"
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

func TestHumblet(t *testing.T) {
	checkLogOutput()

	for n := 2; n <= 100; n++ {
		nodes, runner := lib.BuildCompleteGraph(n)
		clique.RunHumblet(nodes, runner)
	}
}

func BenchmarkHumblet(b *testing.B) {
	b.StopTimer()
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		rand.Seed(0)
		nodes, runner := lib.BuildCompleteGraph(100)

		b.StartTimer()
		clique.RunHumblet(nodes, runner)
		b.StopTimer()
	}
}
