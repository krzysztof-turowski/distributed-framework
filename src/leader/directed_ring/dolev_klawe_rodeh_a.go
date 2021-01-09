package directed_ring

import (
	"encoding/json"
	"fmt"
	"lib"
	"log"
)

type statusDovelKlaweRodeh int

const (
	maximum statusDovelKlaweRodeh = 0
	active  statusDovelKlaweRodeh = 1
	passive statusDovelKlaweRodeh = 2
)

type stateDovelKlaweRodehA struct {
	MaxV   int
	LeftV  int
	Status statusDovelKlaweRodeh
}

type messageTypeDovelKlaweRodeh int

const (
	m1        messageTypeDovelKlaweRodeh = 1
	m2        messageTypeDovelKlaweRodeh = 2
	terminate messageTypeDovelKlaweRodeh = 3 //util type
)

type messageDovelKlaweRodehA struct {
	MessageType messageTypeDovelKlaweRodeh
	Value       int
}

func sendDovelKlaweRodehA(v lib.Node, message messageDovelKlaweRodehA) {
	messageBytes, _ := json.Marshal(message)
	v.SendMessage(0, messageBytes)
}

func receiveDovelKlaweRodehA(v lib.Node) (stateDovelKlaweRodehA, messageDovelKlaweRodehA) {
	var s stateDovelKlaweRodehA
	json.Unmarshal(v.GetState(), &s)

	var m messageDovelKlaweRodehA
	message := v.ReceiveMessage(0)
	json.Unmarshal(message, &m)

	return s, m
}

func initializeDovelKlaweRodehA(v lib.Node) bool {
	s := stateDovelKlaweRodehA{v.GetIndex(), -1, active}

	data, _ := json.Marshal(s)
	v.SetState(data)

	sendDovelKlaweRodehA(v, messageDovelKlaweRodehA{1, s.MaxV})
	return false
}

func processDovelKlaweRodehA(v lib.Node, round int) bool {
	s, m := receiveDovelKlaweRodehA(v)

	if m.MessageType == terminate {
		if s.Status == maximum {
			return true
		}
		sendDovelKlaweRodehA(v, m)
		return true
	}

	if s.Status == passive {
		sendDovelKlaweRodehA(v, m)
	} else if s.Status == active {
		if m.MessageType == m1 {
			if m.Value != s.MaxV {
				sendDovelKlaweRodehA(v, messageDovelKlaweRodehA{m2, m.Value})
				s.LeftV = m.Value
			} else {
				s.Status = maximum
				sendDovelKlaweRodehA(v, messageDovelKlaweRodehA{terminate, 0}) // stop message
			}
		} else if m.MessageType == m2 {
			if s.LeftV > m.Value && s.LeftV > s.MaxV {
				s.MaxV = s.LeftV
				sendDovelKlaweRodehA(v, messageDovelKlaweRodehA{m1, s.MaxV})
			} else {
				s.Status = passive
			}
		}
		data, _ := json.Marshal(s)
		v.SetState(data)
	}
	return false
}

func runDovelKlaweRodehA(v lib.Node) {
	v.StartProcessing()
	finish := initializeDovelKlaweRodehA(v)
	v.FinishProcessing(finish)

	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = processDovelKlaweRodehA(v, round)
		v.FinishProcessing(finish)
	}
}

func checkDovelKlaweRodehA(vertices []lib.Node) {
	var lead_node lib.Node
	var s stateDovelKlaweRodehA

	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if s.Status == maximum {
			lead_node = v
			break
		}
	}

	if lead_node == nil {
		panic("There is no leader on the directed ring")
	}

	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if v != lead_node {

			if s.Status == maximum {
				panic(fmt.Sprint(
					"Multiple leaders on the directed ring: ", lead_node.GetIndex(), v.GetIndex()))
			}

			if s.Status != passive {
				panic(fmt.Sprint("Node ", v.GetIndex(), " has state ", s.Status))
			}
		}
	}
}

func RunDovelKlaweRodehA(n int) (int, int) {
	vertices, synchronizer := lib.BuildSynchronizedDirectedRing(n)

	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runDovelKlaweRodehA(v)
	}
	synchronizer.Synchronize(0)
	checkDovelKlaweRodehA(vertices)

	return synchronizer.GetStats()
}
