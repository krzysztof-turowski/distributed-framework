package sync_phase_king

import (
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"math/rand"
)

/* FAULTY BEHAVIOURS FACTORIES */

func EachMessageRandom(N int, T int) func(lib.Node) bool {
	var cnt = make([]int, N)
	return func(v lib.Node) bool {
		idx := v.GetIndex()
		cnt[idx]++
		if cnt[idx] > 1 {
			receive(v)
		}
		if cnt[idx] > 3*(T+1) {
			return true
		} else {
			for i := 0; i < v.GetOutChannelsCount(); i++ {
				send(v, &Message{V: rand.Intn(3)}, i)
			}
			return false
		}
	}
}
