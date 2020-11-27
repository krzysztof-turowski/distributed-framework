package undirected_graph

import (
	"encoding/json"
	"lib"
	"log"
)

type phaseType string

const (
	setup  phaseType = "setup"
	yoDown           = "yoDown"
	yoUp             = "yoUp"
)

type statusType string

const (
	source   statusType = "source"
	sink                = "sink"
	internal            = "internal"
)

type stateYoYo struct {
	// Orientation[i] == 0 if edge to the i-th neighbour is ingoing, 1 if outgoing. Maybe enum instead of bool
	Orientation []bool
	Phase       phaseType
	Status      statusType
}

type messageYoYo struct {
	Source int
}

func getState(v lib.Node) stateYoYo {
	var s stateYoYo
	json.Unmarshal(v.GetState(), &s)
	return s
}

func setState(v lib.Node, s stateYoYo) {
	data, _ := json.Marshal(s)
	v.SetState(data)
}

func receiveYoYo(v lib.Node, i int) messageYoYo {
	var m messageYoYo
	inMessage := v.ReceiveMessage(i)
	json.Unmarshal(inMessage, &m)
	return m
}

func initializeYoYo(v lib.Node) bool {
	s := stateYoYo{Orientation: make([]bool, v.GetOutChannelsCount()), Phase: setup}
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		outMessage, _ := json.Marshal(messageYoYo{Source: v.GetIndex()})
		v.SendMessage(i, outMessage)
	}
	setState(v, s)
	return false
}

func processSetup(v lib.Node, s *stateYoYo) {
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		m := receiveYoYo(v, i)
		s.Orientation[i] = v.GetIndex() < m.Source
	}

	hasOutEdge, hasInEdge := false, false
	for i := 0; i < v.GetOutChannelsCount() && !(hasOutEdge && hasInEdge); i++ {
		if s.Orientation[i] {
			hasOutEdge = true
		} else {
			hasInEdge = true
		}
	}

	if hasInEdge && hasOutEdge {
		s.Status = internal
	} else if hasInEdge {
		s.Status = sink
	} else if hasOutEdge {
		s.Status = source
	} else {
		log.Panic("Graph contains a detached vertex")
	}
}

func processYoYo(v lib.Node, round int) bool {
	s := getState(v)
	switch s.Phase {
	case setup:
		processSetup(v, &s)
	case yoUp:

	case yoDown:
	}
	setState(v, s)
	return true
}

func runYoYo(v lib.Node) {
	v.StartProcessing()
	finish := initializeYoYo(v)
	v.FinishProcessing(finish)

	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = processYoYo(v, round)
		v.FinishProcessing(finish)
	}
}

func checkYoYo(vertices []lib.Node) {
	for _, v := range vertices {
		s := getState(v)
		for i := 0; i < v.GetOutChannelsCount(); i++ {
			log.Printf("(vertex %d) orientation of edge %d is %t\n", v.GetIndex(), i, s.Orientation[i])
		}
		log.Printf("(vertex %d) %s\n", v.GetIndex(), s.Status)
	}
}

func RunYoYo(n int) {
	vertices, synchronizer := lib.BuildSynchronizedRandomGraph(n, float64(0.25))
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runYoYo(v)
	}
	synchronizer.Synchronize(0)
	synchronizer.GetStats()
	checkYoYo(vertices)
}
