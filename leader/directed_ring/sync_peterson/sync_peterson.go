package sync_peterson

import (
	"encoding/json"
	"fmt"
	. "github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/common"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
)

type state struct {
	Max    int
	Status ModeType
}

type message struct {
	Max  int
	Mode ModeType
}

func initialize(v lib.Node) bool {
	s := state{Max: v.GetIndex(), Status: Unknown}
	data, _ := json.Marshal(s)
	v.SetState(data)
	return false
}

func sendMessage(v lib.Node, m message) {
	outMessage, _ := json.Marshal(m)
	v.SendMessage(0, outMessage)
}

func receiveMessage(v lib.Node) message {
	var m message
	inMessage := v.ReceiveMessage(0)
	json.Unmarshal(inMessage, &m)
	return m
}

func getState(v lib.Node) state {
	var s state
	json.Unmarshal(v.GetState(), &s)
	return s
}

func setState(v lib.Node, s state) {
	data, _ := json.Marshal(s)
	v.SetState(data)
}

func process(v lib.Node, round int) bool {
	s := getState(v)
	if s.Status == Unknown {
		sendMessage(v, message{Max: s.Max, Mode: Unknown})
		m1 := receiveMessage(v)
		if m1.Mode == Leader {
			s.Max = m1.Max
			s.Status = NonLeader
			sendMessage(v, message{Max: m1.Max, Mode: Leader})
			setState(v, s)
			return true
		} else if m1.Max == v.GetIndex() {
			s.Status = Leader
			sendMessage(v, message{Max: s.Max, Mode: Leader})
		} else {
			if s.Max > m1.Max {
				sendMessage(v, message{Max: s.Max, Mode: Unknown})
			} else {
				sendMessage(v, message{Max: m1.Max, Mode: Unknown})
			}
			m2 := receiveMessage(v)
			if m2.Mode == Leader {
				s.Max = m2.Max
				s.Status = NonLeader
				sendMessage(v, message{Max: m2.Max, Mode: Leader})
				setState(v, s)
				return true
			} else if m2.Max == v.GetIndex() {
				s.Status = Leader
				sendMessage(v, message{Max: s.Max, Mode: Leader})
			} else if m1.Max >= s.Max && m1.Max >= m2.Max {
				s.Max = m1.Max
			} else {
				s.Status = Relay
			}
		}
		setState(v, s)
	} else if s.Status == Leader {
		m := receiveMessage(v)
		if m.Mode == Leader {
			return true
		}
	} else if s.Status == Relay {
		for i := 0; i < 2; i++ {
			m := receiveMessage(v)
			s.Max = m.Max
			if m.Mode == Leader {
				s.Status = NonLeader
				setState(v, s)
				sendMessage(v, m)
				return true
			} else if m.Max == v.GetIndex() {
				s.Status = Leader
				setState(v, s)
				sendMessage(v, message{Max: m.Max, Mode: Leader})
				break
			} else {
				sendMessage(v, m)
			}
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
		s = getState(v)
		if s.Status == Leader {
			leadNode = v
			break
		}
	}
	if leadNode == nil {
		panic("There is no Leader on the directed undirected_ring")
	}
	for _, v := range vertices {
		s = getState(v)
		if v == leadNode {
			if v.GetIndex() != s.Max {
				panic(fmt.Sprint("Leader has index ", v.GetIndex(), " but maximum ", s.Max))
			}
		} else {
			if s.Status == Leader {
				panic(fmt.Sprint(
					"Multiple Leaders on the directed undirected_ring: ", leadNode.GetIndex(), v.GetIndex()))
			}
			if s.Status != NonLeader {
				panic(fmt.Sprint("Node ", v.GetIndex(), " has state ", s.Status))
			}
			if leadNode.GetIndex() != s.Max {
				panic(fmt.Sprint(
					"Leader has index ", leadNode.GetIndex(), " but node ", v.GetIndex(),
					" has maximum ", s.Max))
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
