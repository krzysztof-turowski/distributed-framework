package sync_ben_or

import (
	"encoding/json"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
	"math/rand"
)

/* TYPES */

type ExchangeRound int

const (
	ER0     ExchangeRound = iota // start to 1st send
	ER1                          // 1st receive to 2nd send
	ER2                          // 2nd receive to 1st send
	ER1Done                      // 1st receive to 2nd send (without decision change)
	ER2Done                      // 2nd receive to end (without decision change)
)

type State struct {
	N             int
	T             int
	ExchangeRound ExchangeRound
	V             int
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
	s.ExchangeRound = ER0
	s.V = V
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
		if i != node.GetIndex()-1 {
			send(node, msg, i)
		} else {
			sendEmpty(node, i)
		}
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

	var C [3]int
	for k := 0; k <= 1; k++ {
		C[k] = countMessages(s.Msgs, k)
	}

	if 2*C[0] > s.N+s.T {
		broadcast(node, &Message{V: 0})
	} else if 2*C[1] > s.N+s.T {
		broadcast(node, &Message{V: 1})
	} else {
		broadcastEmpty(node)
	}

	s.ExchangeRound = ER2
}

func er2(node lib.Node, s *State) {
	s.Msgs = receive(node)

	var D [3]int
	for k := 0; k <= 1; k++ {
		D[k] = countMessages(s.Msgs, k)
	}

	done := false
	if D[0] >= s.T+1 || D[1] >= s.T+1 {
		if D[0] >= s.T+1 {
			s.V = 0
		} else if D[1] >= s.T+1 {
			s.V = 1
		}
		if 2*(D[0]+D[1]) > s.N+s.T {
			done = true
		}
	} else {
		s.V = rand.Intn(2)
	}

	log.Println("Processor", node.GetIndex(), "about to finish phase with V =", s.V)

	broadcast(node, &Message{V: s.V})

	if done {
		s.ExchangeRound = ER1Done
	} else {
		s.ExchangeRound = ER1
	}
}

func er1done(node lib.Node, s *State) {
	node.IgnoreFutureMessages()
	log.Println("Node", node.GetIndex(), "ignores incoming messages from now")

	broadcast(node, &Message{V: s.V})
	s.ExchangeRound = ER2Done
}

func er2done(node lib.Node, s *State) {
	broadcast(node, &Message{V: s.V})
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
	case ER1Done:
		er1done(node, s)
	case ER2Done:
		er2done(node, s)
		finish = true
	}
	setState(node, s)

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
