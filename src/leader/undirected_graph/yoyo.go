package undirected_graph

import (
	"lib"
	"log"
)

func receiveMessageYoDown(v lib.Node, s *stateYoYo, index int) int {
	message := receiveMessageYoYo(v, index)
	s.LastMessages[index] = message.Content
	return message.Content
}

func receiveMessageYoUp(v lib.Node, s *stateYoYo, index int) (bool, bool) {
	message := receiveMessageYoYo(v, index)
	vote, pruneRequested := toBools(message.Content)

	if pruneRequested == 1 {
		s.EdgeStates[index] = pruned
	}

	return vote == 1, pruneRequested == 1
}

func sendMessageYoUp(v lib.Node, index int, answer bool, pruneRequested bool) {
	outMessage := messageYoYo{Content: toInt(answer, pruneRequested)}
	sendMessageYoYo(v, index, outMessage)
}

func initializeYoYo(v lib.Node) bool {
	if v.GetOutChannelsCount() == 0 {
		log.Panic("Graph contains a node adjacent to 0 edges")
	}

	s := newStateYoYo(v.GetOutChannelsCount())
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		sendMessageYoYo(v, i, messageYoYo{Content: v.GetIndex()})
	}
	setStateYoYo(v, s)
	return false
}

func processSetup(v lib.Node, s *stateYoYo) {
	log.Printf("Vertex %d processing setup phase\n", v.GetIndex())
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		m := receiveMessageYoYo(v, i)
		if v.GetIndex() == m.Content {
			log.Panic("Graph contains adjacent vertices with equal ID")
		}
		b := v.GetIndex() < m.Content
		if b {
			log.Printf("Vertex %d orienting its edge %d as outgoing\n", v.GetIndex(), i)
			s.EdgeStates[i] = out
		} else {
			log.Printf("Vertex %d orienting its edge %d as ingoing\n", v.GetIndex(), i)
			s.EdgeStates[i] = in
		}
	}

	s.UpdateStatus()
	log.Printf("Vertex %d updated status is %s\n", v.GetIndex(), s.Status)
}

func processYoDown(v lib.Node, s *stateYoYo) {
	log.Printf("Vertex %d (%s) processing YO- phase\n", v.GetIndex(), s.Status)
	s.Min = v.GetIndex()
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		if s.EdgeStates[i] == in {
			value := receiveMessageYoDown(v, s, i)
			if value < s.Min {
				s.Min = value
			}
		}
	}

	outMessage := messageYoYo{Content: s.Min}
	log.Printf("Vertex %d (%s) passing its minimum %d to neighbors\n", v.GetIndex(), s.Status, s.Min)
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		if s.EdgeStates[i] == out {
			sendMessageYoYo(v, i, outMessage)
		}
	}
}

func processYoUp(v lib.Node, s *stateYoYo) bool {
	log.Printf("Vertex %d (%s) processing -YO phase\n", v.GetIndex(), s.Status)
	finalVote := true
	needsFlipping := make([]bool, v.GetOutChannelsCount())
	outEdgesLeft := 0
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		if s.EdgeStates[i] == out {
			vote, pruned := receiveMessageYoUp(v, s, i)
			finalVote = finalVote && vote
			needsFlipping[i] = !vote
			if !pruned {
				outEdgesLeft++
			}
		}
	}

	requestPrune, pruneDefault := s.PreprocessPruning(outEdgesLeft)
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		if s.EdgeStates[i] == in {
			vote := finalVote && (s.Min == s.LastMessages[i])
			if pruneDefault || requestPrune[i] {
				log.Printf("Vertex %d (%s) at edge %d votes %t and request pruning\n", v.GetIndex(), s.Status, i, vote)
			} else {
				log.Printf("Vertex %d (%s) at edge %d votes %t\n", v.GetIndex(), s.Status, i, vote)
			}
			sendMessageYoUp(v, i, vote, pruneDefault || requestPrune[i])
			needsFlipping[i] = !vote
			if pruneDefault || requestPrune[i] {
				s.EdgeStates[i] = pruned
			}
		}
	}

	for i := 0; i < v.GetOutChannelsCount(); i++ {
		if needsFlipping[i] {
			s.FlipEdge(i)
		}
	}

	isLeader := s.UpdateStatus()
	log.Printf("Vertex %d updated status is %s\n", v.GetIndex(), s.Status)
	return isLeader
}

func processYoYo(v lib.Node, round int) bool {
	s := getStateYoYo(v)
	finish := false
	switch s.Phase {
	case setup:
		processSetup(v, s)
		s.Phase = yoDown
	case yoDown:
		processYoDown(v, s)
		s.Phase = yoUp
	case yoUp:
		finish = processYoUp(v, s)
		s.Phase = yoDown
	}
	setStateYoYo(v, s)
	return finish || s.Status == detached
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
	min := vertices[0].GetIndex()
	leader := -1
	for _, v := range vertices {
		s := getStateYoYo(v)
		if s.Status != detached {
			log.Printf("Leader is %d\n", v.GetIndex())
			leader = v.GetIndex()
		}
		if min > v.GetIndex() {
			min = v.GetIndex()
		}
	}
	log.Printf("Minimum ID was %d\n", min)
	if min != leader {
		panic("Algorithm's output is wrong")
	}
}

func RunYoYo(vertices []lib.Node, synchronizer lib.Synchronizer) {
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runYoYo(v)
	}
	synchronizer.Synchronize(0)
	synchronizer.GetStats()
	checkYoYo(vertices)
}

func RunYoYoRandom(n int, p float64) {
	RunYoYo(lib.BuildSynchronizedRandomGraph(n, p))
}
