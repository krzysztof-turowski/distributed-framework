package test

import (
	"github.com/krzysztof-turowski/distributed-framework/leader/clique/async_humblet"
	"github.com/krzysztof-turowski/distributed-framework/leader/clique/async_loui_matsushita_west"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"io/ioutil"
	"log"
	"math/rand"
	"testing"
)

func TestHumblet(t *testing.T) {
	checkLogOutput()

	for n := 2; n <= 100; n++ {
		nodes, runner := lib.BuildCompleteGraph(n)
		async_humblet.Run(nodes, runner)
	}
}

func TestLouiMatsushitaWest(t *testing.T) {
	checkLogOutput()

	for n := 2; n <= 100; n++ {
		nodes, runner := lib.BuildHamiltonianOrientedCompleteGraph(n)
		async_loui_matsushita_west.Run(nodes, runner)
	}
}

func BenchmarkHumblet(b *testing.B) {
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		rand.Seed(0)
		nodes, runner := lib.BuildCompleteGraph(100)
		async_humblet.Run(nodes, runner)
	}
}

func BenchmarkLouiMatsushitaWest(b *testing.B) {
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		rand.Seed(0)
		nodes, runner := lib.BuildHamiltonianOrientedCompleteGraph(100)
		async_loui_matsushita_west.Run(nodes, runner)
	}
}
