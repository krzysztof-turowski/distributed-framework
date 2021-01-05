package directed_ring

import (
	"encoding/json"
	"fmt"
	"lib"
	"log"
)

type statusDovelKlaweRodeh int
const (
	maximum  statusDovelKlaweRodeh = 0
	active  statusDovelKlaweRodeh = 1
	passive statusDovelKlaweRodeh = 2
)

type stateDovelKlaweRodeh struct {
	MaxV   int64
	LeftV  int64
	Status statusDovelKlaweRodeh
}

type messageTypeDovelKlaweRodeh int

type messageDovelKlaweRodeh struct {
	MessageType messageTypeDovelKlaweRodeh
	Value       int64
}

func sendDovelKlaweRodeh(v lib.Node, message messageDovelKlaweRodeh) {
	messageBytes, _ := json.Marshal(message)
	v.SendMessage(0, messageBytes)
}

func receiveDovelKlaweRodeh(v lib.Node) (stateDovelKlaweRodeh, messageDovelKlaweRodeh) {
	var s stateDovelKlaweRodeh
	json.Unmarshal(v.GetState(), &s)

	var m messageDovelKlaweRodeh
	message := v.ReceiveMessage(0)
	json.Unmarshal(message, &m)
	return s, m
}

func initializeDovelKlaweRodeh(v lib.Node) bool {
	s := stateDovelKlaweRodeh{int64(v.GetIndex()), -1, active}

	data, _ := json.Marshal(s)
	v.SetState(data)

	sendDovelKlaweRodeh(v, messageDovelKlaweRodeh{1, s.MaxV})
	return false
}


func processDovelKlaweRodeh(v lib.Node, round int) bool {
	s, m := receiveDovelKlaweRodeh(v)

	if m.MessageType == 0 {
		if s.Status == maximum {
			return true
		}
		sendDovelKlaweRodeh(v, m)
		return true
	}

	if s.Status == passive {
		sendDovelKlaweRodeh(v, m)
	} else if s.Status == active {

		if m.MessageType == 1 {

			if m.Value != s.MaxV {
				sendDovelKlaweRodeh(v, messageDovelKlaweRodeh{2,m.Value})
				s.LeftV = m.Value
			} else {
				s.Status = maximum
				sendDovelKlaweRodeh(v, messageDovelKlaweRodeh{0,0}) // stop message
			}
		} else if m.MessageType == 2 {

			if s.LeftV > m.Value && s.LeftV > s.MaxV{
				s.MaxV = s.LeftV
				sendDovelKlaweRodeh(v, messageDovelKlaweRodeh{1, s.MaxV})
			} else {
				s.Status = passive
			}
		}

		data, _ := json.Marshal(s)
		v.SetState(data)
	}

	return false
}

func runDovelKlaweRodeh(v lib.Node) {
	v.StartProcessing()
	finish := initializeDovelKlaweRodeh(v)
	v.FinishProcessing(finish)

	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = processDovelKlaweRodeh(v, round)
		v.FinishProcessing(finish)
	}
}


func checkDovelKlaweRodeh(vertices []lib.Node) {
	var lead_node lib.Node
	var s stateDovelKlaweRodeh

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
				panic(fmt.Sprintf(
					"Multiple leaders on the directed ring:", lead_node.GetIndex(), v.GetIndex()))
			}

			if s.Status != passive {
				panic(fmt.Sprintf("Node", v.GetIndex(), "has state", s.Status))
			}
		}
	}

}

func RunDovelKlaweRodeh(n int)  {
	vertices, synchronizer := lib.BuildSynchronizedDirectedRing(n)

	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runDovelKlaweRodeh(v)
	}
	synchronizer.Synchronize(0)
	synchronizer.GetStats()
	checkDovelKlaweRodeh(vertices)
}


