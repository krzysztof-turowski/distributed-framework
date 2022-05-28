package sync_single_bit

import (
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"math/rand"
)

type Random struct {
}

func (r *Random) er0(v lib.Node, _ []lib.Node, _ map[int]int, _ []*Message) {
	broadcastRandomMsgs(v)
}

func (r *Random) er1(v lib.Node, _ []lib.Node, _ map[int]int, _ []*Message) {
	broadcastRandomMsgs(v)
}

func broadcastRandomMsgs(v lib.Node) {
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		send(v, &Message{V: rand.Intn(2)}, i)
	}
}
