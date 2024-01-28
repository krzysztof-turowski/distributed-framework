package test

import (
	"github.com/krzysztof-turowski/distributed-framework/leader/clique/async_loui_matsushita_west_2"
	"io/ioutil"
	"log"
	"math/rand"
	"testing"

	"github.com/krzysztof-turowski/distributed-framework/leader/clique/async_afek_gafni_b"
	"github.com/krzysztof-turowski/distributed-framework/leader/clique/async_humblet"
	"github.com/krzysztof-turowski/distributed-framework/leader/clique/async_korach_moran_zaks"
	"github.com/krzysztof-turowski/distributed-framework/leader/clique/async_loui_matsushita_west"
	"github.com/krzysztof-turowski/distributed-framework/lib"
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

func TestKorachMoranZaks(t *testing.T) {
	checkLogOutput()
	for n := 2; n <= 100; n++ {
		nodes, runner := lib.BuildCompleteGraph(n)
		async_korach_moran_zaks.Run(nodes, runner)
	}
}

func TestAfekGafniB(t *testing.T) {
	checkLogOutput()
	for n := 2; n <= 100; n++ {
		nodes, runner := lib.BuildCompleteGraph(n)
		async_afek_gafni_b.Run(nodes, runner)
	}
}

func TestLouiMatsushitaWest2(t *testing.T) {
	checkLogOutput()
	for n := 2; n <= 100; n++ {
		nodes, runner := lib.BuildCompleteGraph(n)
		async_loui_matsushita_west_2.Run(nodes, runner)
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

func BenchmarkKorachMoranZaks(b *testing.B) {
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		rand.Seed(0)
		nodes, runner := lib.BuildCompleteGraph(100)
		async_korach_moran_zaks.Run(nodes, runner)
	}
}

func BenchmarkAfekGafniB(b *testing.B) {
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		rand.Seed(0)
		nodes, runner := lib.BuildCompleteGraph(100)
		async_afek_gafni_b.Run(nodes, runner)
	}
}

func BenchmarkLouiMatsushitaWest2(b *testing.B) {
	log.SetOutput(ioutil.Discard)

	for i := 0; i < b.N; i++ {
		rand.Seed(0)
		nodes, runner := lib.BuildCompleteGraph(100)
		async_loui_matsushita_west_2.Run(nodes, runner)
	}
}
