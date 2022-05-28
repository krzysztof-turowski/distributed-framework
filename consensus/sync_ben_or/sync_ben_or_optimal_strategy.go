package sync_ben_or

import (
	"encoding/json"
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

type Optimal struct {
}

func (r *Optimal) er0(node lib.Node, nodes []lib.Node, faultyIndices map[int]int, _ []*Message) {
	vals := peekV(nodes)
	var E int
	for _, val := range vals {
		if !val.null && val.value == 0 {
			E++
		}
	}
	if j := faultyIndices[node.GetIndex()]; j+E <= (len(nodes)+len(faultyIndices))/2 {
		broadcast(node, &Message{V: 0})
	} else {
		broadcast(node, &Message{V: 1})
	}
}

func (r *Optimal) er1(node lib.Node, _ []lib.Node, _ map[int]int, _ []*Message) []*Message {
	broadcastEmpty(node)
	return make([]*Message, node.GetOutChannelsCount())
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
