package sync_single_bit

import (
	"encoding/json"
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

type Optimal struct {
}

func (r *Optimal) er0(v lib.Node, nodes []lib.Node, faultyIndices map[int]int, _ []*Message) {
	vals := peekV(nodes)
	var E int
	for _, val := range vals {
		if !val.null && val.value == 0 {
			E++
		}
	}
	if j := faultyIndices[v.GetIndex()]; 4*(j+E) < 3*len(nodes) {
		broadcast(v, &Message{V: 0})
	} else {
		broadcast(v, &Message{V: 1})
	}
}

func (r *Optimal) er1(v lib.Node, _ []lib.Node, faultyIndices map[int]int, _ []*Message) {
	for i := 0; i < len(faultyIndices)+1; i++ {
		send(v, &Message{V: 0}, i)
	}
	for i := len(faultyIndices) + 1; i < v.GetOutChannelsCount(); i++ {
		send(v, &Message{V: 1}, i)
	}
}

type maybeInt struct {
	value int
	null  bool
}

func peekV(nodes []lib.Node) []maybeInt {
	allV := make([]maybeInt, len(nodes))
	for i, node := range nodes {
		if node.GetState() == nil {
			allV[i] = maybeInt{null: true}
		} else {
			var s State
			json.Unmarshal(node.GetState(), &s)
			allV[i] = maybeInt{value: s.V, null: false}
		}
	}
	return allV
}
