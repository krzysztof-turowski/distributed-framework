package sync_kuhn_wattenhofer

import (
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
)

func runSynchronized(vertices []lib.Node, synchronizer lib.Synchronizer, runNode func(lib.Node), check func([]lib.Node)) {
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runNode(v)
	}
	synchronizer.Synchronize(0)
	synchronizer.GetStats()
	check(vertices)
}

func run(v lib.Node, init, process func(lib.Node) bool) {
	v.StartProcessing()
	finish := init(v)
	v.FinishProcessing(finish)

	for !finish {
		v.StartProcessing()
		finish = process(v)
		v.FinishProcessing(finish)
	}
}
