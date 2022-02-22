package sync_dolev_klawe_rodeh

import (
	"encoding/json"
	"fmt"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
)

type status int

const (
	maximum status = 0
	active  status = 1
	passive status = 2
)

type stateA struct {
	MaxV   int
	LeftV  int
	Status status
}

type messageType int

const (
	m1        messageType = 1
	m2        messageType = 2
	terminate messageType = 3 //util type
)

type messageA struct {
	MessageType messageType
	Value       int
}

func sendA(v lib.Node, message messageA) {
	messageBytes, _ := json.Marshal(message)
	v.SendMessage(0, messageBytes)
}

func receiveA(v lib.Node) (stateA, messageA) {
	var s stateA
	json.Unmarshal(v.GetState(), &s)

	var m messageA
	message := v.ReceiveMessage(0)
	json.Unmarshal(message, &m)

	return s, m
}

func initializeA(v lib.Node) bool {
	s := stateA{v.GetIndex(), -1, active}

	data, _ := json.Marshal(s)
	v.SetState(data)

	sendA(v, messageA{1, s.MaxV})
	return false
}

func processA(v lib.Node, round int) bool {
	s, m := receiveA(v)

	if m.MessageType == terminate {
		if s.Status == maximum {
			return true
		}
		sendA(v, m)
		return true
	}

	if s.Status == passive {
		sendA(v, m)
	} else if s.Status == active {
		if m.MessageType == m1 {
			if m.Value != s.MaxV {
				sendA(v, messageA{m2, m.Value})
				s.LeftV = m.Value
			} else {
				s.Status = maximum
				sendA(v, messageA{terminate, 0}) // stop message
			}
		} else if m.MessageType == m2 {
			if s.LeftV > m.Value && s.LeftV > s.MaxV {
				s.MaxV = s.LeftV
				sendA(v, messageA{m1, s.MaxV})
			} else {
				s.Status = passive
			}
		}
		data, _ := json.Marshal(s)
		v.SetState(data)
	}
	return false
}

func runA(v lib.Node) {
	v.StartProcessing()
	finish := initializeA(v)
	v.FinishProcessing(finish)

	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = processA(v, round)
		v.FinishProcessing(finish)
	}
}

func checkA(vertices []lib.Node) {
	var lead_node lib.Node
	var s stateA

	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if s.Status == maximum {
			lead_node = v
			break
		}
	}

	if lead_node == nil {
		panic("There is no leader on the directed undirected_ring")
	}

	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if v != lead_node {

			if s.Status == maximum {
				panic(fmt.Sprint(
					"Multiple leaders on the directed undirected_ring: ", lead_node.GetIndex(), v.GetIndex()))
			}

			if s.Status != passive {
				panic(fmt.Sprint("Node ", v.GetIndex(), " has state ", s.Status))
			}
		}
	}
}

func RunA(n int) (int, int) {
	vertices, synchronizer := lib.BuildSynchronizedDirectedRing(n)

	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runA(v)
	}
	synchronizer.Synchronize(0)
	checkA(vertices)

	return synchronizer.GetStats()
}
