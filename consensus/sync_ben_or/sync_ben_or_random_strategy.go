package sync_ben_or

import (
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"math/rand"
)

type Random struct {
}

func (r *Random) er0(node lib.Node, _ []lib.Node, _ map[int]int, _ []*Message) {
	broadcastRandomMsgs(node)
}

func (r *Random) er1(node lib.Node, _ []lib.Node, _ map[int]int, _ []*Message) []*Message {
	return broadcastRandomOrEmptyMsgs(node)
}

func broadcastRandomMsgs(node lib.Node) {
	for i := 0; i < node.GetOutChannelsCount(); i++ {
		if i == node.GetIndex()-1 {
			sendEmpty(node, i)
		} else {
			send(node, &Message{V: rand.Intn(2)}, i)
		}
	}
}

func broadcastRandomOrEmptyMsgs(node lib.Node) []*Message {
	msgs := make([]*Message, node.GetOutChannelsCount())
	for i := 0; i < node.GetOutChannelsCount(); i++ {
		r := rand.Intn(3)
		if r == 2 || i == node.GetIndex()-1 {
			sendEmpty(node, i)
			msgs[i] = nil
		} else {
			send(node, &Message{V: r}, i)
			msgs[i] = &Message{V: r}
		}
	}
	return msgs
}
