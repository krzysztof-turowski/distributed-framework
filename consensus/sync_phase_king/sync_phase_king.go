package sync_phase_king

import (
	"encoding/json"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
)

/* TYPES */

type ExchangeRound int

const (
	ER0 ExchangeRound = iota // start to 1st send
	ER1                      // 1st receive to 2nd send
	ER2                      // 2nd receive to 3rd send
	ER3                      // 3rd receive to 1st send
)

type State struct {
	N             int
	T             int
	Phase         int
	ExchangeRound ExchangeRound
	V             int
	D             [3]int
	Msgs          []*Message
}

type Message struct {
	V int
}

/* STATE METHODS */

func newState(n int, t int, V int) *State {
	var s State
	s.N = n
	s.T = t
	s.Phase = 1
	s.ExchangeRound = ER0
	s.V = V
	s.D = [3]int{}
	return &s
}

func getState(node lib.Node) *State {
	var s State
	json.Unmarshal(node.GetState(), &s)
	return &s
}

func setState(node lib.Node, s *State) {
	sAsJson, _ := json.Marshal(*s)
	node.SetState(sAsJson)
}

/* MESSAGE METHODS */

func send(node lib.Node, msg *Message, dest int) {
	msgAsJson, _ := json.Marshal(msg)
	node.SendMessage(dest, msgAsJson)
}

func sendEmpty(node lib.Node, dest int) {
	node.SendMessage(dest, nil)
}

func broadcast(node lib.Node, msg *Message) {
	for i := 0; i < node.GetOutChannelsCount(); i++ {
		send(node, msg, i)
	}
}

func broadcastEmpty(node lib.Node) {
	for i := 0; i < node.GetOutChannelsCount(); i++ {
		sendEmpty(node, i)
	}
}

func receive(node lib.Node) []*Message {
	msgs := make([]*Message, node.GetInChannelsCount())

	for i := 0; i < node.GetInChannelsCount(); i++ {
		msgAsJson := node.ReceiveMessage(i)
		if msgAsJson != nil {
			msgs[i] = &Message{}
			json.Unmarshal(msgAsJson, msgs[i])
		}
	}
	return msgs
}

/* METHODS */

func countMessages(msgs []*Message, k int) int {
	r := 0
	for _, msg := range msgs {
		if msg != nil && msg.V == k {
			r++
		}
	}
	return r
}

func er0(node lib.Node, s *State) {
	broadcast(node, &Message{V: s.V})
	s.ExchangeRound = ER1
}

func er1(node lib.Node, s *State) {
	s.Msgs = receive(node)
	s.V = 2
	var C [3]int
	for k := 0; k <= 1; k++ {
		C[k] = countMessages(s.Msgs, k)
		if C[k] >= s.N-s.T {
			s.V = k
		}
	}
	broadcast(node, &Message{V: s.V})
	s.ExchangeRound = ER2
}

func er2(node lib.Node, s *State) {
	s.Msgs = receive(node)
	for k := 2; k >= 0; k-- {
		s.D[k] = countMessages(s.Msgs, k)
		if s.D[k] > s.T {
			s.V = k
		}
	}
	if s.Phase == node.GetIndex() {
		broadcast(node, &Message{V: s.V})
	} else {
		broadcastEmpty(node)
	}
	s.ExchangeRound = ER3
}

func er3(node lib.Node, s *State) bool {
	s.Msgs = receive(node)
	kingMsg := s.Msgs[s.Phase-1]

	if s.V == 2 || s.D[s.V] < s.N-s.T {
		if kingMsg != nil && kingMsg.V == 0 {
			s.V = 0
		} else {
			s.V = 1
		}
	}

	s.Phase++
	if s.Phase > s.T+1 {
		return true
	}

	broadcast(node, &Message{V: s.V})
	s.ExchangeRound = ER1
	return false
}

func process(node lib.Node) bool {
	s := getState(node)
	finish := false
	switch s.ExchangeRound {
	case ER0:
		er0(node, s)
	case ER1:
		er1(node, s)
	case ER2:
		er2(node, s)
	case ER3:
		finish = er3(node, s)
	}
	setState(node, s)
	log.Println("Processor", node.GetIndex(), "about to finish er", s.ExchangeRound, "with V =", s.V)

	return finish
}

func run(node lib.Node, N int, T int, V int) {
	setState(node, newState(N, T, V))

	for finish := false; !finish; {
		node.StartProcessing()
		finish = process(node)
		node.FinishProcessing(finish)
	}
}

func runFaulty(node lib.Node, faultyBehavior func(lib.Node) bool) {
	for finish := false; !finish; {
		node.StartProcessing()
		finish = faultyBehavior(node)
		node.FinishProcessing(finish)
	}
}

func check(nodes []lib.Node, V []int, faultyIndices map[int]int) int {
	var consensus int
	for i, node := range nodes {
		if _, ok := faultyIndices[i+1]; !ok {
			consensus = getState(node).V
			break
		}
	}

	for i, node := range nodes {
		if _, ok := faultyIndices[i+1]; !ok && getState(node).V != consensus {
			panic("Agreement not reached")
		}
	}
	for i, v := range V {
		if _, ok := faultyIndices[i+1]; !ok && v == consensus {
			return consensus
		}
	}
	panic("Agreement not valid")
}

func Run(
	nodes []lib.Node,
	synchronizer lib.Synchronizer,
	V []int,
	faultyBehavior func(lib.Node) bool,
	faultyIndices map[int]int,
) (int, int) {

	for i, node := range nodes {
		if _, ok := faultyIndices[i+1]; !ok {
			log.Println("Correct processor", node.GetIndex(), "about to run")
			go run(node, len(nodes), len(faultyIndices), V[i])
		} else {
			log.Println("Faulty processor", node.GetIndex(), "about to run")
			go runFaulty(node, faultyBehavior)
		}
	}
	synchronizer.Synchronize(0)
	log.Println("Correct processors agreed on", check(nodes, V, faultyIndices))
	return synchronizer.GetStats()
}
