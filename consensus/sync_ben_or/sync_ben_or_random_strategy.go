package sync_ben_or

import (
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"math/rand"
)

type Random struct {
}

func (r *Random) er0(v lib.Node, _ []lib.Node, _ map[int]int, _ []*Message) {
	broadcastRandomMsgs(v)
}

func (r *Random) er1(v lib.Node, _ []lib.Node, _ map[int]int, _ []*Message) []*Message {
	return broadcastRandomOrEmptyMsgs(v)
}

func broadcastRandomMsgs(v lib.Node) {
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		if i == v.GetIndex()-1 {
			sendEmpty(v, i)
		} else {
			send(v, &Message{V: rand.Intn(2)}, i)
		}
	}
}

func broadcastRandomOrEmptyMsgs(v lib.Node) []*Message {
	msgs := make([]*Message, v.GetOutChannelsCount())
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		r := rand.Intn(3)
		if r == 2 || i == v.GetIndex()-1 {
			sendEmpty(v, i)
			msgs[i] = nil
		} else {
			send(v, &Message{V: r}, i)
			msgs[i] = &Message{V: r}
		}
	}
	return msgs
}
