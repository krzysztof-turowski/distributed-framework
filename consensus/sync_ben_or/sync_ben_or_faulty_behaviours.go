package sync_ben_or

import (
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
	"math/rand"
)

/* FAULTY BEHAVIOURS FACTORIES */

// EachMessageRandom
// Faulty processors share common data. They assume having indices from 1 to T.
// First of them is responsible for finding out when to finish.
func EachMessageRandom(N int, T int) func(lib.Node) bool {

	type State struct {
		invocation      int
		fakeDecidedMsgs int
		exchangeRound   ExchangeRound
	}
	var states []*State
	for i := 0; i < N; i++ {
		states = append(states, &State{invocation: 0, fakeDecidedMsgs: 0, exchangeRound: ER0})
	}
	someoneIsDone := false

	return func(v lib.Node) bool {
		s := states[v.GetIndex()]
		s.invocation++

		if s.exchangeRound == ER1 && someoneIsDone {
			s.exchangeRound = ER1Done
		}

		switch s.exchangeRound {
		case ER0:
			broadcastRandomMsgs(v)
			s.exchangeRound = ER1
		case ER1:
			receive(v)
			for i := 0; i < v.GetOutChannelsCount(); i++ {
				if rand.Intn(2) == 0 {
					send(v, &Message{V: rand.Intn(2)}, i)
					states[i].fakeDecidedMsgs++
				} else {
					sendEmpty(v, i)
				}
			}
			s.exchangeRound = ER2
		case ER2:
			msgs := receive(v)
			if v.GetIndex() == 1 {
				realDecidedMsgs := 0
				for i := T; i < v.GetInChannelsCount(); i++ {
					if msgs[i] != nil {
						realDecidedMsgs++
					}
				}
				for i := T; i < v.GetOutChannelsCount(); i++ {
					if 2*(realDecidedMsgs+states[i].fakeDecidedMsgs) > N+T {
						someoneIsDone = true
						break
					}
				}
				for i := 0; i < v.GetOutChannelsCount(); i++ {
					states[i].fakeDecidedMsgs = 0
				}
			}
			broadcastRandomMsgs(v)
			s.exchangeRound = ER1
		case ER1Done:
			v.IgnoreFutureMessages()
			log.Println("Node", v.GetIndex(), "ignores incoming messages from now")

			broadcastRandomMsgs(v)
			s.exchangeRound = ER2Done
		case ER2Done:
			broadcastRandomMsgs(v)
			return true
		}

		return false
	}
}

func broadcastRandomMsgs(v lib.Node) {
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		send(v, &Message{V: rand.Intn(2)}, i)
	}
}
