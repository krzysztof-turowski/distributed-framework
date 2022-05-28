package sync_phase_king

import (
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"math/rand"
)

type Random struct {
}

func (r *Random) er0(v lib.Node, _ []lib.Node, _ map[int]int, _ []*Message) {
	broadcastRandomMsgs(v, 2)
}

func (r *Random) er1(v lib.Node, _ []lib.Node, _ map[int]int, _ []*Message) {
	broadcastRandomMsgs(v, 3)
}

func (r *Random) er2(v lib.Node, _ []lib.Node, _ map[int]int, _ []*Message) {
	broadcastRandomMsgs(v, 2)
}

func broadcastRandomMsgs(v lib.Node, r int) {
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		send(v, &Message{V: rand.Intn(r)}, i)
	}
}
