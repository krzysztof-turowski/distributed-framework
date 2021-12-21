package directed_ring

import (
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"encoding/json"
	"fmt"
	"log"
)

type statusDolevKlaweRodeh int

const (
	maximum statusDolevKlaweRodeh = 0
	active  statusDolevKlaweRodeh = 1
	passive statusDolevKlaweRodeh = 2
)

type stateDolevKlaweRodehA struct {
	MaxV   int
	LeftV  int
	Status statusDolevKlaweRodeh
}

type messageTypeDolevKlaweRodeh int

const (
	m1        messageTypeDolevKlaweRodeh = 1
	m2        messageTypeDolevKlaweRodeh = 2
	terminate messageTypeDolevKlaweRodeh = 3 //util type
)

type messageDolevKlaweRodehA struct {
	MessageType messageTypeDolevKlaweRodeh
	Value       int
}

func sendDolevKlaweRodehA(v lib.Node, message messageDolevKlaweRodehA) {
	messageBytes, _ := json.Marshal(message)
	v.SendMessage(0, messageBytes)
}

func receiveDolevKlaweRodehA(v lib.Node) (stateDolevKlaweRodehA, messageDolevKlaweRodehA) {
	var s stateDolevKlaweRodehA
	json.Unmarshal(v.GetState(), &s)

	var m messageDolevKlaweRodehA
	message := v.ReceiveMessage(0)
	json.Unmarshal(message, &m)

	return s, m
}

func initializeDolevKlaweRodehA(v lib.Node) bool {
	s := stateDolevKlaweRodehA{v.GetIndex(), -1, active}

	data, _ := json.Marshal(s)
	v.SetState(data)

	sendDolevKlaweRodehA(v, messageDolevKlaweRodehA{1, s.MaxV})
	return false
}

func processDolevKlaweRodehA(v lib.Node, round int) bool {
	s, m := receiveDolevKlaweRodehA(v)

	if m.MessageType == terminate {
		if s.Status == maximum {
			return true
		}
		sendDolevKlaweRodehA(v, m)
		return true
	}

	if s.Status == passive {
		sendDolevKlaweRodehA(v, m)
	} else if s.Status == active {
		if m.MessageType == m1 {
			if m.Value != s.MaxV {
				sendDolevKlaweRodehA(v, messageDolevKlaweRodehA{m2, m.Value})
				s.LeftV = m.Value
			} else {
				s.Status = maximum
				sendDolevKlaweRodehA(v, messageDolevKlaweRodehA{terminate, 0}) // stop message
			}
		} else if m.MessageType == m2 {
			if s.LeftV > m.Value && s.LeftV > s.MaxV {
				s.MaxV = s.LeftV
				sendDolevKlaweRodehA(v, messageDolevKlaweRodehA{m1, s.MaxV})
			} else {
				s.Status = passive
			}
		}
		data, _ := json.Marshal(s)
		v.SetState(data)
	}
	return false
}

func runDolevKlaweRodehA(v lib.Node) {
	v.StartProcessing()
	finish := initializeDolevKlaweRodehA(v)
	v.FinishProcessing(finish)

	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = processDolevKlaweRodehA(v, round)
		v.FinishProcessing(finish)
	}
}

func checkDolevKlaweRodehA(vertices []lib.Node) {
	var lead_node lib.Node
	var s stateDolevKlaweRodehA

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

func RunDolevKlaweRodehA(n int) (int, int) {
	vertices, synchronizer := lib.BuildSynchronizedDirectedRing(n)

	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runDolevKlaweRodehA(v)
	}
	synchronizer.Synchronize(0)
	checkDolevKlaweRodehA(vertices)

	return synchronizer.GetStats()
}
