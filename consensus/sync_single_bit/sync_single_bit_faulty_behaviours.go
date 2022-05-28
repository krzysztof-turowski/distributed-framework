package sync_single_bit

import (
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

func GetFaultyBehaviour(nodes []lib.Node, faultyIndices map[int]int, strategy Strategy) func(lib.Node) bool {

	type State struct {
		phase         int
		exchangeRound ExchangeRound
	}
	states := make([]*State, len(nodes))
	for i := 0; i < len(nodes); i++ {
		states[i] = &State{phase: 1, exchangeRound: ER0}
	}

	return func(v lib.Node) bool {
		s := states[v.GetIndex()-1]

		switch s.exchangeRound {
		case ER0:
			strategy.er0(v, nodes, faultyIndices, nil)
			s.exchangeRound = ER1
		case ER1:
			msgs := receive(v)
			if s.phase == v.GetIndex() {
				strategy.er1(v, nodes, faultyIndices, msgs)
			} else {
				broadcastEmpty(v)
			}
			s.exchangeRound = ER2
		case ER2:
			msgs := receive(v)
			s.phase++
			if s.phase > len(faultyIndices)+1 {
				return true
			}
			strategy.er0(v, nodes, faultyIndices, msgs)
			s.exchangeRound = ER1
		}

		return false
	}
}

type Strategy interface {
	er0(v lib.Node, nodes []lib.Node, faultyIndices map[int]int, msgs []*Message)
	er1(v lib.Node, nodes []lib.Node, faultyIndices map[int]int, msgs []*Message)
}
