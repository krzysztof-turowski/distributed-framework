package sync_single_bit

import (
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

func GetFaultyBehavior(nodes []lib.Node, faultyIndices map[int]int, strategy Strategy) func(lib.Node) bool {

	type State struct {
		phase         int
		exchangeRound ExchangeRound
	}
	states := make([]*State, len(nodes))
	for i := 0; i < len(nodes); i++ {
		states[i] = &State{phase: 1, exchangeRound: ER0}
	}

	return func(node lib.Node) bool {
		s := states[node.GetIndex()-1]

		switch s.exchangeRound {
		case ER0:
			strategy.er0(node, nodes, faultyIndices, nil)
			s.exchangeRound = ER1
		case ER1:
			msgs := receive(node)
			if s.phase == node.GetIndex() {
				strategy.er1(node, nodes, faultyIndices, msgs)
			} else {
				broadcastEmpty(node)
			}
			s.exchangeRound = ER2
		case ER2:
			msgs := receive(node)
			s.phase++
			if s.phase > len(faultyIndices)+1 {
				return true
			}
			strategy.er0(node, nodes, faultyIndices, msgs)
			s.exchangeRound = ER1
		}

		return false
	}
}

type Strategy interface {
	er0(node lib.Node, nodes []lib.Node, faultyIndices map[int]int, msgs []*Message)
	er1(node lib.Node, nodes []lib.Node, faultyIndices map[int]int, msgs []*Message)
}
