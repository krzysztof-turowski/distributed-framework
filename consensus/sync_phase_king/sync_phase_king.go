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
		send(v, msg, i)
	}
}

func broadcastEmpty(v lib.Node) {
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		sendEmpty(v, i)
	}
}

func receive(v lib.Node) []*Message {
	var msgs []*Message

	for i := 0; i < v.GetInChannelsCount(); i++ {
		msgAsJson := v.ReceiveMessage(i)
		msgs = append(msgs, &Message{})
		if msgAsJson != nil {
			json.Unmarshal(msgAsJson, msgs[i])
		} else {
			msgs[i] = nil
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
	s.V = 2
	var C [3]int
	for k := 0; k <= 1; k++ {
		C[k] = countMessages(s.Msgs, k)
		if C[k] >= s.N-s.T {
			s.V = k
		}
	}
	broadcast(v, &Message{V: s.V})
	s.ExchangeRound = ER2
}

func er2(v lib.Node, s *State) {
	s.Msgs = receive(v)
	for k := 2; k >= 0; k-- {
		s.D[k] = countMessages(s.Msgs, k)
		if s.D[k] > s.T {
			s.V = k
		}
	}
	if s.Phase == v.GetIndex() {
		broadcast(v, &Message{V: s.V})
	} else {
		broadcastEmpty(v)
	}
	s.ExchangeRound = ER3
}

func er3(v lib.Node, s *State) bool {
	s.Msgs = receive(v)
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

	broadcast(v, &Message{V: s.V})
	s.ExchangeRound = ER1
	return false
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
	case ER3:
		finish = er3(v, s)
	}
	setState(v, s)

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

func check(vertices []lib.Node, V []int) int {
	consensus := getState(vertices[0]).V
	for _, v := range vertices[1:] {
		if getState(v).V != consensus {
			panic("Agreement not reached")
		}
	}
	for _, v := range V {
		if v == consensus {
			return consensus
		}
	}
	panic("Agreement not valid")
}

func Run(vertices []lib.Node, synchronizer lib.Synchronizer, T int, V []int, faultyBehaviour func(lib.Node) bool) (int, int) {
	N := len(vertices)
	for i, v := range vertices {
		if i >= T {
			log.Println("Correct processor", v.GetIndex(), "about to run")
			go run(v, N, T, V[i])
		} else {
			log.Println("Faulty processor", v.GetIndex(), "about to run")
			go runFaulty(v, faultyBehaviour)
		}
	}
	synchronizer.Synchronize(0)
	log.Println("Correct processors agreed on", check(vertices[T:], V[T:]))
	return synchronizer.GetStats()
}
