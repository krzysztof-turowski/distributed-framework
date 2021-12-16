package directed_hypercube

import (
	"lib"
	"log"
)

type StatusType int

const (
	duellist StatusType = iota
	defeated
	leader
	follower
)

type MessageType int

const (
	match MessageType = iota
	follow
)

type MessageHyperelect struct {
	Type        MessageType
	Value       int
	Stage       int
	Source      []bool
	Destination []int
}

type StateHypercube struct {
	Status       StatusType
	Stage        int
	Delayed      []MessageHyperelect
	NextDuellist []int
}

func initializeHypercube(v lib.Node) bool {
	// TODO: complete init

	return false
}

func hyperElect(v lib.Node) bool {
	// TODO: complete process

	return false
}

func runHyperElect(v lib.Node) {
	v.StartProcessing()
	finish := initializeHypercube(v)
	v.FinishProcessing(finish)

	for !finish {
		v.StartProcessing()
		finish = hyperElect(v)
		v.FinishProcessing(finish)
	}
}

func checkHyperElect(vertices []lib.Node) {
	// TODO: complete check
}

func RunHyperElect(vertices []lib.Node, synchronizer lib.Synchronizer) {
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runHyperElect(v)
	}
	synchronizer.Synchronize(0)
	synchronizer.GetStats()
	checkHyperElect(vertices)
}
