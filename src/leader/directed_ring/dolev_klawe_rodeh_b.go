package directed_ring

import (
	"encoding/json"
	"fmt"
	"lib"
	"log"
)

const (
	waiting statusDovelKlaweRodeh = 3
)

type stateDovelKlaweRodehB struct {
	PhaseV int
	MaxV   int
	Status statusDovelKlaweRodeh
}


type messageDovelKlaweRodehB struct {
	MessageType messageTypeDovelKlaweRodeh
	Value       int
	Phase       int
	Counter     int
}

func sendDovelKlaweRodehB(v lib.Node, message interface{}) {
	if message != nil {
		messageBytes, _ := json.Marshal(message)
		v.SendMessage(0, messageBytes)
	}else {
		v.SendMessage(0, nil)
	}
}

func receiveDovelKlaweRodehB(v lib.Node) (stateDovelKlaweRodehB, messageDovelKlaweRodehB) {
	var s stateDovelKlaweRodehB
	json.Unmarshal(v.GetState(), &s)

	var m messageDovelKlaweRodehB
	message := v.ReceiveMessage(0)
	json.Unmarshal(message, &m)
	return s, m
}

func initializeDovelKlaweRodehB(v lib.Node) bool {
	s := stateDovelKlaweRodehB{0, v.GetIndex(), active}

	data, _ := json.Marshal(s)
	v.SetState(data)

	sendDovelKlaweRodehB(v, messageDovelKlaweRodehB{1, s.MaxV, s.PhaseV, 1})
	return false
}


func processDovelKlaweRodehB(v lib.Node, round int) bool {
	s, m := receiveDovelKlaweRodehB(v)

	if m.MessageType == 0 {
		if s.Status != maximum {
			sendDovelKlaweRodehB(v, nil)
		}
		return false
	}

	if m.Value == v.GetIndex() {
		s.Status = maximum
		data, _ := json.Marshal(s)
		v.SetState(data)
		sendDovelKlaweRodehB(v, messageDovelKlaweRodehB{3, 0, 0,0})
		return false
	}

	if m.MessageType == 3 {
		if s.Status == maximum {
			return true
		}
		sendDovelKlaweRodehB(v, m)
		return true
	}

	if s.Status == active {

		if m.MessageType == 1 {

			if m.Value > s.MaxV {
				s.MaxV = m.Value
				s.Status = waiting
				sendDovelKlaweRodehB(v, nil) // nil
			} else {
				s.Status = passive
				sendDovelKlaweRodehB(v, messageDovelKlaweRodehB{2, s.MaxV, 0,0})
			}

		} else if m.MessageType == 2 {
			//never receive
		}

	}  else if s.Status == waiting {

		if m.MessageType == 1 {
			s.Status = passive

			if m.Value >= s.MaxV && m.Counter >= 1 {
				s.MaxV = m.Value

				if m.Counter > 1 {
					sendDovelKlaweRodehB(v, messageDovelKlaweRodehB{1, m.Value, m.Phase,m.Counter-1})
				} else if m.Counter == 1 {
					s.Status = waiting
					sendDovelKlaweRodehB(v, messageDovelKlaweRodehB{1, m.Value, m.Phase,0})
					s.PhaseV = m.Phase
				}

			} else {
				sendDovelKlaweRodehB(v, messageDovelKlaweRodehB{1, m.Value, m.Phase,0})
			}


		} else if m.MessageType == 2 {
			s.PhaseV = s.PhaseV + 1
			s.Status = active
			sendDovelKlaweRodehB(v, messageDovelKlaweRodehB{1, s.MaxV, s.PhaseV, 1})

		}
	} else if s.Status == passive {

		if m.MessageType == 1 {

			if m.Value >= s.MaxV && m.Counter >= 1 {
				s.MaxV = m.Value

				if m.Counter > 1 {
					sendDovelKlaweRodehB(v, messageDovelKlaweRodehB{1, m.Value, m.Phase,m.Counter-1})
				} else if m.Counter == 1 {
					s.Status = waiting
					sendDovelKlaweRodehB(v, messageDovelKlaweRodehB{1, m.Value, m.Phase,0})
					s.PhaseV = m.Phase
				}

			} else {
				sendDovelKlaweRodehB(v, messageDovelKlaweRodehB{1, m.Value, m.Phase,0})
			}
		} else if m.MessageType == 2 {

			if m.Value < s.MaxV {
				sendDovelKlaweRodehB(v, nil)
			}  else {
				sendDovelKlaweRodehB(v, messageDovelKlaweRodehB{2, m.Value, 0,0})
			}
		}

	}

	data, _ := json.Marshal(s)
	v.SetState(data)

	return false
}

func runDovelKlaweRodehB(v lib.Node) {
	v.StartProcessing()
	finish := initializeDovelKlaweRodehB(v)
	v.FinishProcessing(finish)

	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = processDovelKlaweRodehB(v, round)
		v.FinishProcessing(finish)
	}
}


func checkDovelKlaweRodehB(vertices []lib.Node) {
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
		}
	}

}

func RunDovelKlaweRodehB(n int)  {
	vertices, synchronizer := lib.BuildSynchronizedDirectedRing(n)

	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runDovelKlaweRodehB(v)
	}
	synchronizer.Synchronize(0)
	synchronizer.GetStats()
	checkDovelKlaweRodehB(vertices)
}
