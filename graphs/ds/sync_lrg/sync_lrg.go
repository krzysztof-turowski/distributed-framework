package sync_lrg

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"sync/atomic"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

const B_PARAMETER = 2.0

type state struct {
	Selected bool
	Covered  bool
}

type message struct {
	Value int
}

func setState(v lib.Node, s state) {
	data, _ := json.Marshal(s)
	v.SetState(data)
}

func getState(v lib.Node) state {
	data := v.GetState()
	var res state
	json.Unmarshal(data, &res)
	return res
}

func initialize(v lib.Node) bool {
	setState(v, state{Selected: false, Covered: false})
	return false
}

func sendToAllNeighbors(v lib.Node, value int) {
	mess, _ := json.Marshal(message{Value: value})
	for i := 0; i < v.GetInChannelsCount(); i++ {
		go v.SendMessage(i, mess)
	}
}

func sendNullToNeigbors(v lib.Node) {
	for i := 0; i < v.GetInChannelsCount(); i++ {
		go v.SendMessage(i, nil)
	}
}

func receive(v lib.Node, index int) int {
	mess := v.ReceiveMessage(index)
	var res message
	json.Unmarshal(mess, &res)
	return res.Value
}

func process(v lib.Node, done *int32) {
	uncovered_nei := make([]bool, v.GetInChannelsCount())
	mystate := getState(v)
	covered := 0
	if mystate.Covered {
		covered = 1
	}
	selected := mystate.Selected
	span := 0
	calcSpan := func() {
		for i := 0; i < v.GetInChannelsCount(); i++ {
			uncovered_nei[i] = false
		}
		sendToAllNeighbors(v, covered)
		span = 1 - covered
		for i := 0; i < v.GetInChannelsCount(); i++ {
			tmp := receive(v, i)
			span += 1 - tmp
			uncovered_nei[i] = tmp == 0
		}
	}

	calcRoundedSpan := func() int {
		res := 1.0
		cnt := 0
		for ; res < float64(span); cnt++ {
			res *= B_PARAMETER
		}
		return cnt
	}

	getMaxFromNei := func() int {
		res := 0
		for i := 0; i < v.GetInChannelsCount(); i++ {
			tmp := receive(v, i)
			if tmp > res {
				res = tmp
			}
		}
		return res
	}

	getSumFromNei := func() int {
		res := 0
		for i := 0; i < v.GetInChannelsCount(); i++ {
			res += receive(v, i)
		}
		return res
	}

	getMedian := func(me int) int {
		res := make([]int, span)
		cnt := 0
		if covered == 0 {
			res[cnt] = me
			cnt++
		}
		for i := 0; i < v.GetInChannelsCount(); i++ {
			if uncovered_nei[i] {
				res[cnt] = receive(v, i)
				cnt++
			} else {
				v.ReceiveMessage(i) // nil
			}
		}
		if span == 0 {
			return 0
		}
		sort.Ints(res)
		return res[cnt/2]
	}
	// calculating span, i.e. uncovered neighborhood
	calcSpan()
	v.FinishProcessing(false)
	// rounded span
	d_v := calcRoundedSpan()
	// mx = max of span from N^1(v)
	v.StartProcessing()
	sendToAllNeighbors(v, d_v)
	mx := getMaxFromNei()
	v.FinishProcessing(false)
	// mx = max of span from N^2(v)
	v.StartProcessing()
	if d_v > mx {
		mx = d_v
	}
	sendToAllNeighbors(v, mx)
	mx = getMaxFromNei()
	candidate := d_v >= mx && !selected && span != 0
	v.FinishProcessing(false)
	// announce candidates, calculate for every uncovered node s(v) = how many candidates cover it
	v.StartProcessing()
	if candidate {
		sendToAllNeighbors(v, 1)
	} else {
		sendToAllNeighbors(v, 0)
	}
	s_v := getSumFromNei()
	if covered == 0 && candidate {
		s_v++
	}
	v.FinishProcessing(false)
	// gather s(v) from neighbors
	v.StartProcessing()
	if covered == 0 {
		sendToAllNeighbors(v, s_v)
	} else {
		sendNullToNeigbors(v)
	}
	med := getMedian(s_v)
	v.FinishProcessing(false)
	// add candidate to dominating set with chance 1/(median of s_v from uncovered neighbors)
	if med != 0 && candidate {
		if 1/float64(med) >= rand.Float64() {
			selected = true
			setState(v, state{Selected: selected, Covered: true})
		}
	}
	// recalc covered nodes
	v.StartProcessing()
	if selected {
		sendToAllNeighbors(v, 1)
	} else {
		sendToAllNeighbors(v, 0)
	}
	cov_sum := getSumFromNei()
	if selected {
		cov_sum++
	}
	// first time covered nodes contribute to stop condition
	if cov_sum != 0 && covered == 0 {
		atomic.AddInt32(done, -1)
		setState(v, state{Selected: selected, Covered: true})
	}
}

func run(v lib.Node, finish *int32) {
	v.StartProcessing()
	initialize(v)
	v.FinishProcessing(false)
	for {
		v.StartProcessing()
		process(v, finish)
		v.FinishProcessing(false)
		v.StartProcessing()
		if *finish == 0 {
			v.FinishProcessing(true)
			break
		} else {
			v.FinishProcessing(false)
		}
	}
}

func check(vertices []lib.Node) {
	selected := 0
	for _, v := range vertices {
		if !getState(v).Covered {
			panic(fmt.Sprint("Node is not covered"))
		}
		if getState(v).Selected {
			selected++
		}
	}
	log.Println("Selected", selected, "nodes to dominating set")
}

func Run(n int, p float64) (int, int) {
	vertices, synchronizer := lib.BuildSynchronizedRandomGraph(n, p)
	done := int32(n)
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go run(v, &done)
	}

	synchronizer.Synchronize(0)
	check(vertices)
	return synchronizer.GetStats()
}
