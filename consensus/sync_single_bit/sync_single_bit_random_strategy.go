package sync_single_bit

import (
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"math/rand"
)

type Random struct {
}

func (r *Random) er0(node lib.Node, _ []lib.Node, _ map[int]int, _ []*Message) {
	broadcastRandomMsgs(node)
}

func (r *Random) er1(node lib.Node, _ []lib.Node, _ map[int]int, _ []*Message) {
	broadcastRandomMsgs(node)
}

func broadcastRandomMsgs(node lib.Node) {
	for i := 0; i < node.GetOutChannelsCount(); i++ {
		send(node, &Message{V: rand.Intn(2)}, i)
	}
}
