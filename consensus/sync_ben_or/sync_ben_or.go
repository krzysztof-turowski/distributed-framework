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

func getState(v lib.Node) *State {
	var s State
	json.Unmarshal(v.GetState(), &s)
	return &s
}

func setState(v lib.Node, s *State) {
	sAsJson, _ := json.Marshal(*s)
	v.SetState(sAsJson)
}

/* MESSAGE METHODS */

func send(v lib.Node, msg *Message, dest int) {
	msgAsJson, _ := json.Marshal(msg)
	v.SendMessage(dest, msgAsJson)
}

func sendEmpty(v lib.Node, dest int) {
	v.SendMessage(dest, nil)
}

func broadcast(v lib.Node, msg *Message) {
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		if i != v.GetIndex()-1 {
			send(v, msg, i)
		} else {
			sendEmpty(v, i)
		}
	}
}

func broadcastEmpty(v lib.Node) {
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		sendEmpty(v, i)
	}
}

func receive(v lib.Node) []*Message {
	msgs := make([]*Message, v.GetInChannelsCount())

	for i := 0; i < v.GetInChannelsCount(); i++ {
		msgAsJson := v.ReceiveMessage(i)
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

func er0(v lib.Node, s *State) {
	broadcast(v, &Message{V: s.V})
	s.ExchangeRound = ER1
}

func er1(v lib.Node, s *State) {
	s.Msgs = receive(v)

	var C [3]int
	for k := 0; k <= 1; k++ {
		C[k] = countMessages(s.Msgs, k)
	}

	if 2*C[0] > s.N+s.T {
		broadcast(v, &Message{V: 0})
	} else if 2*C[1] > s.N+s.T {
		broadcast(v, &Message{V: 1})
	} else {
		broadcastEmpty(v)
	}

	s.ExchangeRound = ER2
}

func er2(v lib.Node, s *State) {
	s.Msgs = receive(v)

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

	broadcast(v, &Message{V: s.V})

	if done {
		s.ExchangeRound = ER1Done
	} else {
		s.ExchangeRound = ER1
	}
}

func er1done(v lib.Node, s *State) {
	v.IgnoreFutureMessages()
	log.Println("Node", v.GetIndex(), "ignores incoming messages from now")

	broadcast(v, &Message{V: s.V})
	s.ExchangeRound = ER2Done
}

func er2done(v lib.Node, s *State) {
	broadcast(v, &Message{V: s.V})
}

func process(v lib.Node) bool {
	s := getState(v)
	finish := false
	switch s.ExchangeRound {
	case ER0:
		er0(v, s)
	case ER1:
		er1(v, s)
	case ER2:
		er2(v, s)
	case ER1Done:
		er1done(v, s)
	case ER2Done:
		er2done(v, s)
		finish = true
	}
	setState(v, s)
	log.Println("Node", v.GetIndex(), "is about to finish ER =", s.ExchangeRound, "with V =", s.V)

	return finish
}

func run(v lib.Node, N int, T int, V int) {
	setState(v, newState(N, T, V))

	for finish := false; !finish; {
		v.StartProcessing()
		finish = process(v)
		v.FinishProcessing(finish)
	}
}

func runFaulty(v lib.Node, faultyBehaviour func(lib.Node) bool) {
	for finish := false; !finish; {
		v.StartProcessing()
		finish = faultyBehaviour(v)
		v.FinishProcessing(finish)
	}
}

func check(nodes []lib.Node, V []int, faultyIndices map[int]int) int {
	var consensus int
	for i, v := range nodes {
		if _, ok := faultyIndices[i+1]; !ok {
			consensus = getState(v).V
			break
		}
	}

	for i, v := range nodes {
		if _, ok := faultyIndices[i+1]; !ok && getState(v).V != consensus {
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
	faultyBehaviour func(lib.Node) bool,
	faultyIndices map[int]int,
) (int, int) {

	for i, v := range nodes {
		if _, ok := faultyIndices[i+1]; !ok {
			log.Println("Correct processor", v.GetIndex(), "about to run")
			go run(v, len(nodes), len(faultyIndices), V[i])
		} else {
			log.Println("Faulty processor", v.GetIndex(), "about to run")
			go runFaulty(v, faultyBehaviour)
		}
	}
	synchronizer.Synchronize(0)
	log.Println("Correct processors agreed on", check(nodes, V, faultyIndices))
	return synchronizer.GetStats()
}
