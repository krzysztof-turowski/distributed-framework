package sync_ben_or

import (
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
)

// GetFaultyBehaviour
// Faulty processors share common data.
// First of them is responsible for finding out when to finish.
func GetFaultyBehaviour(nodes []lib.Node, faultyIndices map[int]int, strategy Strategy) func(lib.Node) bool {

	type State struct {
		invocation      int
		fakeDecidedMsgs int
		exchangeRound   ExchangeRound
	}
	states := make([]*State, len(nodes))
	for i := 0; i < len(nodes); i++ {
		states[i] = &State{invocation: 0, fakeDecidedMsgs: 0, exchangeRound: ER0}
	}
	someoneIsDone := false

	return func(v lib.Node) bool {
		s := states[v.GetIndex()-1]
		s.invocation++

		if s.exchangeRound == ER1 && someoneIsDone {
			s.exchangeRound = ER1Done
		}

		switch s.exchangeRound {
		case ER0:
			strategy.er0(v, nodes, faultyIndices, nil)
			s.exchangeRound = ER1
		case ER1:
			for i, msg := range strategy.er1(v, nodes, faultyIndices, receive(v)) {
				if msg != nil {
					states[i].fakeDecidedMsgs++
				}
			}
			s.exchangeRound = ER2
		case ER2:
			msgs := receive(v)
			if faultyIndices[v.GetIndex()] == 1 {
				realDecidedMsgs := 0
				for i := 0; i < len(nodes); i++ {
					if _, ok := faultyIndices[i+1]; !ok {
						if msgs[i] != nil {
							realDecidedMsgs++
						}
					}
				}
				for i := 0; i < len(nodes); i++ {
					if _, ok := faultyIndices[i+1]; !ok {
						if 2*(realDecidedMsgs+states[i].fakeDecidedMsgs) > len(nodes)+len(faultyIndices) {
							someoneIsDone = true
							break
						}
					}
				}
				for i := 0; i < len(nodes); i++ {
					states[i].fakeDecidedMsgs = 0
				}
			}
			strategy.er0(v, nodes, faultyIndices, msgs)
			s.exchangeRound = ER1
		case ER1Done:
			v.IgnoreFutureMessages()
			log.Println("Node", v.GetIndex(), "ignores incoming messages from now")

			strategy.er1(v, nodes, faultyIndices, nil)
			s.exchangeRound = ER2Done
		case ER2Done:
			strategy.er0(v, nodes, faultyIndices, nil)
			return true
		}

		return false
	}
}

type Strategy interface {
	er0(v lib.Node, nodes []lib.Node, faultyIndices map[int]int, msgs []*Message)
	er1(v lib.Node, nodes []lib.Node, faultyIndices map[int]int, msgs []*Message) []*Message
}
