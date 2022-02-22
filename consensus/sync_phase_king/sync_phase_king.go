package sync_phase_king

import (
	"encoding/json"
	"log"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

/* PHASEKING TYPES */

type ExchangeRound int

const (
	ER0 ExchangeRound = iota // start to 1st send
	ER1                      // 1st receive to 2nd send
	ER2                      // 2nd receive to 3rd send
	ER3                      // 3rd receive to 1st send
)

type StatePhaseKing struct {
	N             int
	T             int
	Phase         int
	ExchangeRound ExchangeRound
	V             int
	C             [3]int
	D             [3]int
}

type MessagePhaseKing struct {
	V int
}

/* STATE PHASEKING METHODS */

func newStatePhaseKing(n int, t int, V int) *StatePhaseKing {
	var s StatePhaseKing
	s.N = n
	s.T = t
	s.Phase = 1
	s.ExchangeRound = ER0
	s.V = V
	s.C = [3]int{}
	s.D = [3]int{}
	return &s
}

func getStatePhaseKing(v lib.Node) *StatePhaseKing {
	var s StatePhaseKing
	json.Unmarshal(v.GetState(), &s)
	return &s
}

func setStatePhaseKing(v lib.Node, s *StatePhaseKing) {
	sAsJson, _ := json.Marshal(*s)
	v.SetState(sAsJson)
}

/* MESSAGE PHASEKING METHODS */

func sendPhaseKing(v lib.Node, msg *MessagePhaseKing) {
	msgAsJson, _ := json.Marshal(msg)
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		v.SendMessage(i, msgAsJson)
	}
}

func sendEmptyPhaseKing(v lib.Node) {
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		v.SendMessage(i, nil)
	}
}

func receivePhaseKing(v lib.Node) []*MessagePhaseKing {
	msgs := []*MessagePhaseKing{}

	for i := 0; i < v.GetInChannelsCount(); i++ {
		msgAsJson := v.ReceiveMessage(i)
		msgs = append(msgs, &MessagePhaseKing{})
		json.Unmarshal(msgAsJson, msgs[i])
	}
	return msgs
}

/* PHASEKING METHODS */

func er0PhaseKing(v lib.Node, s *StatePhaseKing) {
	sendPhaseKing(v, &MessagePhaseKing{V: s.V})
	s.ExchangeRound = ER1
}

func countMessagesPhaseKing(msgs []*MessagePhaseKing, k int) int {
	r := 0
	for _, msg := range msgs {
		if msg.V == k {
			r++
		}
	}
	return r
}

func er1PhaseKing(v lib.Node, s *StatePhaseKing) {
	msgs := receivePhaseKing(v)
	s.V = 2
	for k := 0; k <= 1; k++ {
		s.C[k] = countMessagesPhaseKing(msgs, k)
		if s.C[k] >= s.N-s.T {
			s.V = k
		}
	}
	sendPhaseKing(v, &MessagePhaseKing{V: s.V})
	s.ExchangeRound = ER2
}

func er2PhaseKing(v lib.Node, s *StatePhaseKing) {
	msgs := receivePhaseKing(v)
	for k := 2; k >= 0; k-- {
		s.D[k] = countMessagesPhaseKing(msgs, k)
		if s.D[k] > s.T {
			s.V = k
		}
	}
	if s.Phase == v.GetIndex() {
		sendPhaseKing(v, &MessagePhaseKing{V: s.V})
	} else {
		sendEmptyPhaseKing(v)
	}
	s.ExchangeRound = ER3
}

func er3PhaseKing(v lib.Node, s *StatePhaseKing) bool {
	msg := receivePhaseKing(v)[s.Phase-1].V

	if s.V == 2 || s.D[s.V] < s.N-s.T {
		if msg == 0 {
			s.V = 0
		} else {
			s.V = 1
		}
	}

	s.Phase++
	if s.Phase > s.T+1 {
		return true
	}

	sendPhaseKing(v, &MessagePhaseKing{V: s.V})
	s.ExchangeRound = ER1
	return false
}

func processPhaseKing(v lib.Node) bool {
	s := getStatePhaseKing(v)
	finish := false
	switch s.ExchangeRound {
	case ER0:
		er0PhaseKing(v, s)
	case ER1:
		er1PhaseKing(v, s)
	case ER2:
		er2PhaseKing(v, s)
	case ER3:
		finish = er3PhaseKing(v, s)
	}
	setStatePhaseKing(v, s)

	return finish
}

func runPhaseKing(v lib.Node, N int, T int, V int) {
	setStatePhaseKing(v, newStatePhaseKing(N, T, V))

	for finish := false; !finish; {
		v.StartProcessing()
		finish = processPhaseKing(v)
		v.FinishProcessing(finish)
	}
}

func processFaultyPhaseKing(v lib.Node) bool {
	s := getStatePhaseKing(v)
	sendPhaseKing(v, &MessagePhaseKing{V: lib.GetRandomGenerator().Int() % 2})
	receivePhaseKing(v)
	s.Phase++
	setStatePhaseKing(v, s)
	return s.Phase > 3*s.T+3
}

func runFaultyPhaseKing(v lib.Node, N int, T int, V int) {
	setStatePhaseKing(v, newStatePhaseKing(N, T, V))

	for finish := false; !finish; {
		v.StartProcessing()
		finish = processFaultyPhaseKing(v)
		v.FinishProcessing(finish)
	}
}

func checkPhaseKing(vertices []lib.Node, V []int) int {
	consensus := getStatePhaseKing(vertices[0]).V
	for _, v := range vertices[1:] {
		if getStatePhaseKing(v).V != consensus {
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

func RunPhaseKing(vertices []lib.Node, synchronizer lib.Synchronizer, T int, V []int) (int, int) {
	N := len(vertices)
	for i, v := range vertices {
		if i >= T {
			log.Println("Correct processor", v.GetIndex(), "about to run")
			go runPhaseKing(v, N, T, V[i])
		} else {
			log.Println("Faulty processor", v.GetIndex(), "about to run")
			go runFaultyPhaseKing(v, N, T, V[i])
		}
	}
	synchronizer.Synchronize(0)
	log.Println("Correct processors agreed on", checkPhaseKing(vertices[T:], V[T:]))
	return synchronizer.GetStats()
}
