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
	Selected        bool
	Covered         bool
	Span            int
	RoundedSpan     int
	BestRoundedSpan int
	Candidate       bool
	Support         int
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

func getMaxFromNei(v lib.Node) int {
	res := 0
	for i := 0; i < v.GetInChannelsCount(); i++ {
		tmp := receive(v, i)
		if tmp > res {
			res = tmp
		}
	}
	return res
}

func getSumFromNei(v lib.Node) int {
	res := 0
	for i := 0; i < v.GetInChannelsCount(); i++ {
		res += receive(v, i)
	}
	return res
}

func roundSpan(span int) int {
	res := 1.0
	cnt := 0
	for ; res < float64(span); cnt++ {
		res *= B_PARAMETER
	}
	return cnt
}

func process(v lib.Node, done *int32, round int) bool {
	mystate := getState(v)
	switch round % 7 {
	case 0: // span calculation
		covered := 0
		if mystate.Covered {
			covered = 1
		}
		sendToAllNeighbors(v, covered)
		span := 1 - covered
		for i := 0; i < v.GetInChannelsCount(); i++ {
			span += 1 - receive(v, i)
		}
		mystate.Span = span
		mystate.RoundedSpan = roundSpan(span)
	case 1: // getting best rounded span from N^1(v)
		sendToAllNeighbors(v, mystate.RoundedSpan)
		mx := getMaxFromNei(v)
		if mystate.RoundedSpan > mx {
			mx = mystate.RoundedSpan
		}
		mystate.BestRoundedSpan = mx
	case 2: // getting best rounded span from N^2(v) and selecting candidates
		sendToAllNeighbors(v, mystate.BestRoundedSpan)
		mx := getMaxFromNei(v)
		mystate.Candidate = mystate.RoundedSpan >= mx && !mystate.Selected && mystate.Span != 0
	case 3: // calculating support
		if mystate.Candidate {
			sendToAllNeighbors(v, 1)
		} else {
			sendToAllNeighbors(v, 0)
		}
		mystate.Support = getSumFromNei(v)
		if !mystate.Covered && mystate.Candidate {
			mystate.Support++
		}
	case 4: // collecting support and selecting new dominators
		if !mystate.Covered {
			sendToAllNeighbors(v, mystate.Support)
		} else {
			sendNullToNeigbors(v)
		}
		res := make([]int, mystate.Span)
		cnt := 0
		if !mystate.Covered {
			res[cnt] = mystate.Support
			cnt++
		}
		for i := 0; i < v.GetInChannelsCount(); i++ {
			tmp := v.ReceiveMessage(i)
			if tmp != nil {
				var val message
				json.Unmarshal(tmp, &val)
				res[cnt] = val.Value
				cnt++
			}
		}
		if mystate.Span != 0 && mystate.Candidate {
			sort.Ints(res)
			if 1/float64(res[cnt/2]) >= rand.Float64() {
				mystate.Selected = true
			}
		}
	case 5: // announcing new dominators
		cnt := 0
		if mystate.Selected {
			cnt++
			sendToAllNeighbors(v, 1)
		} else {
			sendToAllNeighbors(v, 0)
		}
		cnt += getSumFromNei(v)
		if !mystate.Covered && cnt != 0 {
			atomic.AddInt32(done, -1)
			mystate.Covered = true
		}
	case 6: // synchronization
		if *done == 0 {
			return true
		}
	}
	setState(v, mystate)
	return false
}

func run(v lib.Node, done *int32) {
	v.StartProcessing()
	finish := initialize(v)
	v.FinishProcessing(finish)
	for rnd := 0; !finish; rnd++ {
		v.StartProcessing()
		finish := process(v, done, rnd)
		v.FinishProcessing(finish)
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
