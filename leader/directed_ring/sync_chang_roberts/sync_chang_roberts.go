package sync_chang_roberts

import (
	"encoding/json"
	"fmt"
	. "github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/common"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
)

type state struct {
	Min    int
	Status ModeType
}

type message struct {
	Min  int
	Mode ModeType
}

func send(v lib.Node, s state, mode ModeType) {
	if mode != Pass {
		outMessage, _ := json.Marshal(message{Min: s.Min, Mode: mode})
		v.SendMessage(0, outMessage)
	} else {
		v.SendMessage(0, nil)
	}
	data, _ := json.Marshal(s)
	v.SetState(data)
}

func receive(v lib.Node) (state, message) {
	var s state
	json.Unmarshal(v.GetState(), &s)
	var m message
	inMessage := v.ReceiveMessage(0)
	json.Unmarshal(inMessage, &m)
	return s, m
}

func initialize(v lib.Node) bool {
	s := state{Min: v.GetIndex(), Status: Unknown}
	send(v, s, Unknown)
	return false
}

func process(v lib.Node, round int) bool {
	if s, m := receive(v); s.Status == Unknown {
		if m.Mode == Unknown {
			if m.Min == v.GetIndex() {
				s.Status = Leader
				send(v, s, Leader)
			} else if m.Min < s.Min {
				s.Min = m.Min
				send(v, s, Unknown)
			} else {
				send(v, s, Pass)
			}
		} else if m.Mode == Leader {
			s.Status = NonLeader
			send(v, s, Leader)
			return true
		} else {
			send(v, s, Pass)
		}
	} else if s.Status == Leader {
		if m.Mode == Leader {
			return true
		}
	}
	return false
}

func run(v lib.Node) {
	v.StartProcessing()
	finish := initialize(v)
	v.FinishProcessing(finish)

	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = process(v, round)
		v.FinishProcessing(finish)
	}
}

func check(vertices []lib.Node) {
	var leadNode lib.Node
	var s state
	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if s.Status == Leader {
			leadNode = v
			break
		}
	}
	if leadNode == nil {
		panic("There is no Leader on the directed undirected_ring")
	}
	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if v == leadNode {
			if v.GetIndex() != s.Min {
				panic(fmt.Sprint("Leader has index ", v.GetIndex(), " but minimum ", s.Min))
			}
		} else {
			if s.Status == Leader {
				panic(fmt.Sprint(
					"Multiple Leaders on the directed undirected_ring: ", leadNode.GetIndex(), v.GetIndex()))
			}
			if s.Status != NonLeader {
				panic(fmt.Sprint("Node ", v.GetIndex(), " has state ", s.Status))
			}
			if leadNode.GetIndex() != s.Min {
				panic(fmt.Sprint(
					"Leader has index ", leadNode.GetIndex(), " but node ", v.GetIndex(),
					" has minimum ", s.Min))
			}
		}
	}
}

func Run(n int) (int, int) {
	vertices, synchronizer := lib.BuildSynchronizedDirectedRing(n)
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go run(v)
	}
	synchronizer.Synchronize(0)
	check(vertices)

	return synchronizer.GetStats()
}
