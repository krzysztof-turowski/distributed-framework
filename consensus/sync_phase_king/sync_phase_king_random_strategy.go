package sync_phase_king

import (
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"math/rand"
)

type Random struct {
}

func (r *Random) er0(node lib.Node, _ []lib.Node, _ map[int]int, _ []*Message) {
	broadcastRandomMsgs(node, 2)
}

func (r *Random) er1(node lib.Node, _ []lib.Node, _ map[int]int, _ []*Message) {
	broadcastRandomMsgs(node, 3)
}

func (r *Random) er2(node lib.Node, _ []lib.Node, _ map[int]int, _ []*Message) {
	broadcastRandomMsgs(node, 2)
}

func broadcastRandomMsgs(node lib.Node, r int) {
	for i := 0; i < node.GetOutChannelsCount(); i++ {
		send(node, &Message{V: rand.Intn(r)}, i)
	}
}
