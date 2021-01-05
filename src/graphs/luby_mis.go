package graphs

import (
	"lib"
	"log"
)

type stateLubyMIS struct {
}

type messageLubyMIS struct {
}

func initializeLubyMIS(v lib.Node) bool {
	return true
}

func processLubyMIS(v lib.Node, round int) bool {
	return true
}

func runLubyMIS(v lib.Node) {
	v.StartProcessing()
	finish := initializeLubyMIS(v)
	v.FinishProcessing(finish)

	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = processLubyMIS(v, round)
		v.FinishProcessing(finish)
	}
}

func checkLubyMIS(vertices []lib.Node) {
	panic("The algorithm is not implemented yet, so results cannot be correct")
}

func RunLubyMIS(n int, p float64) {
	vertices, synchronizer := lib.BuildSynchronizedRandomGraph(n, p)
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runLubyMIS(v)
	}

	synchronizer.Synchronize(0)
	synchronizer.GetStats()
	checkLubyMIS(vertices)
}
