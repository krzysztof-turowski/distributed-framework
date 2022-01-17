package directed_ring

import (
	"encoding/json"
	"fmt"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
)

type stateChangRoberts struct {
	Min    int
	Status modeType
}

type messageChangRoberts struct {
	Min  int
	Mode modeType
}

func sendChangRoberts(v lib.Node, s stateChangRoberts, mode modeType) {
	if mode != pass {
		outMessage, _ := json.Marshal(messageChangRoberts{Min: s.Min, Mode: mode})
		v.SendMessage(0, outMessage)
	} else {
		v.SendMessage(0, nil)
	}
	data, _ := json.Marshal(s)
	v.SetState(data)
}

func receiveChangRoberts(v lib.Node) (stateChangRoberts, messageChangRoberts) {
	var s stateChangRoberts
	json.Unmarshal(v.GetState(), &s)
	var m messageChangRoberts
	inMessage := v.ReceiveMessage(0)
	json.Unmarshal(inMessage, &m)
	return s, m
}

func initializeChangRoberts(v lib.Node) bool {
	s := stateChangRoberts{Min: v.GetIndex(), Status: unknown}
	sendChangRoberts(v, s, unknown)
	return false
}

func processChangRoberts(v lib.Node, round int) bool {
	if s, m := receiveChangRoberts(v); s.Status == unknown {
		if m.Mode == unknown {
			if m.Min == v.GetIndex() {
				s.Status = leader
				sendChangRoberts(v, s, leader)
			} else if m.Min < s.Min {
				s.Min = m.Min
				sendChangRoberts(v, s, unknown)
			} else {
				sendChangRoberts(v, s, pass)
			}
		} else if m.Mode == leader {
			s.Status = nonleader
			sendChangRoberts(v, s, leader)
			return true
		} else {
			sendChangRoberts(v, s, pass)
		}
	} else if s.Status == leader {
		if m.Mode == leader {
			return true
		}
	}
	return false
}

func runChangRoberts(v lib.Node) {
	v.StartProcessing()
	finish := initializeChangRoberts(v)
	v.FinishProcessing(finish)

	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = processChangRoberts(v, round)
		v.FinishProcessing(finish)
	}
}

func checkChangRoberts(vertices []lib.Node) {
	var lead_node lib.Node
	var s stateChangRoberts
	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if s.Status == leader {
			lead_node = v
			break
		}
	}
	if lead_node == nil {
		panic("There is no leader on the directed ring")
	}
	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if v == lead_node {
			if v.GetIndex() != s.Min {
				panic(fmt.Sprint("Leader has index ", v.GetIndex(), " but minimum ", s.Min))
			}
		} else {
			if s.Status == leader {
				panic(fmt.Sprint(
					"Multiple leaders on the directed ring: ", lead_node.GetIndex(), v.GetIndex()))
			}
			if s.Status != nonleader {
				panic(fmt.Sprint("Node ", v.GetIndex(), " has state ", s.Status))
			}
			if lead_node.GetIndex() != s.Min {
				panic(fmt.Sprint(
					"Leader has index ", lead_node.GetIndex(), " but node ", v.GetIndex(),
					" has minimum ", s.Min))
			}
		}
	}
}

func RunChangRoberts(n int) (int, int) {
	vertices, synchronizer := lib.BuildSynchronizedDirectedRing(n)
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runChangRoberts(v)
	}
	synchronizer.Synchronize(0)
	checkChangRoberts(vertices)

	return synchronizer.GetStats()
}
