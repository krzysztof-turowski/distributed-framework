package sync_single_bit

import (
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"math/rand"
)

/* FAULTY BEHAVIOURS FACTORIES */

func EachMessageRandom(N int, T int) func(lib.Node) bool {
	var count = make([]int, N)
	return func(v lib.Node) bool {
		index := v.GetIndex()
		count[index]++
		if count[index] > 1 {
			receive(v)
		}
		if count[index] > 2*(T+1) {
			return true
		} else {
			for i := 0; i < v.GetOutChannelsCount(); i++ {
				send(v, &Message{V: rand.Intn(2)}, i)
			}
			return false
		}
	}
}
