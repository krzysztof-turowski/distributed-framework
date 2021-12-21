package directed_ring

import (
	"distributed-framework/lib"
	"encoding/json"
	"fmt"
	"log"
)

const (
	waiting statusDolevKlaweRodeh = 3
)

type stateDolevKlaweRodehB struct {
	PhaseV int
	MaxV   int
	Status statusDolevKlaweRodeh
}

const (
	empty messageTypeDolevKlaweRodeh = 0 //util type
)

type messageDolevKlaweRodehB struct {
	MessageType messageTypeDolevKlaweRodeh
	Value       int
	Phase       int
	Counter     int
}

func sendDolevKlaweRodehB(v lib.Node, message interface{}) {
	if message != nil {
		messageBytes, _ := json.Marshal(message)
		v.SendMessage(0, messageBytes)
	} else {
		v.SendMessage(0, nil)
	}
}

func receiveDolevKlaweRodehB(v lib.Node) (stateDolevKlaweRodehB, messageDolevKlaweRodehB) {
	var s stateDolevKlaweRodehB
	json.Unmarshal(v.GetState(), &s)

	var m messageDolevKlaweRodehB
	message := v.ReceiveMessage(0)
	json.Unmarshal(message, &m)

	return s, m
}

func initializeDolevKlaweRodehB(v lib.Node) bool {
	s := stateDolevKlaweRodehB{0, v.GetIndex(), active}

	data, _ := json.Marshal(s)
	v.SetState(data)

	sendDolevKlaweRodehB(v, messageDolevKlaweRodehB{1, s.MaxV, s.PhaseV, 1})
	return false
}

func processDolevKlaweRodehB(v lib.Node, round int) bool {
	s, m := receiveDolevKlaweRodehB(v)

	if m.MessageType == empty {
		if s.Status != maximum {
			sendDolevKlaweRodehB(v, nil)
		}
		return false
	}

	if m.Value == v.GetIndex() {
		s.Status = maximum
		data, _ := json.Marshal(s)
		v.SetState(data)
		sendDolevKlaweRodehB(v, messageDolevKlaweRodehB{terminate, 0, 0, 0})
		return false
	}

	if m.MessageType == terminate {
		if s.Status == maximum {
			return true
		}
		sendDolevKlaweRodehB(v, m)
		return true
	}

	if s.Status == active {
		if m.MessageType == m1 {
			if m.Value > s.MaxV {
				s.MaxV = m.Value
				s.Status = waiting
				sendDolevKlaweRodehB(v, nil) // nil
			} else {
				s.Status = passive
				sendDolevKlaweRodehB(v, messageDolevKlaweRodehB{m2, s.MaxV, 0, 0})
			}
		} else if m.MessageType == m2 {
			//never receive
		}
	} else if s.Status == waiting {
		if m.MessageType == m1 {
			s.Status = passive
			if m.Value >= s.MaxV && m.Counter >= 1 {
				s.MaxV = m.Value
				if m.Counter > 1 {
					sendDolevKlaweRodehB(v, messageDolevKlaweRodehB{m1, m.Value, m.Phase, m.Counter - 1})
				} else if m.Counter == 1 {
					s.Status = waiting
					sendDolevKlaweRodehB(v, messageDolevKlaweRodehB{m1, m.Value, m.Phase, 0})
					s.PhaseV = m.Phase
				}
			} else {
				sendDolevKlaweRodehB(v, messageDolevKlaweRodehB{m1, m.Value, m.Phase, 0})
			}
		} else if m.MessageType == m2 {
			s.PhaseV = s.PhaseV + 1
			s.Status = active
			sendDolevKlaweRodehB(v, messageDolevKlaweRodehB{m1, s.MaxV, s.PhaseV, 1})

		}
	} else if s.Status == passive {
		if m.MessageType == m1 {
			if m.Value >= s.MaxV && m.Counter >= 1 {
				s.MaxV = m.Value
				if m.Counter > 1 {
					sendDolevKlaweRodehB(v, messageDolevKlaweRodehB{m1, m.Value, m.Phase, m.Counter - 1})
				} else if m.Counter == 1 {
					s.Status = waiting
					sendDolevKlaweRodehB(v, messageDolevKlaweRodehB{m1, m.Value, m.Phase, 0})
					s.PhaseV = m.Phase
				}
			} else {
				sendDolevKlaweRodehB(v, messageDolevKlaweRodehB{m1, m.Value, m.Phase, 0})
			}
		} else if m.MessageType == 2 {
			if m.Value < s.MaxV {
				sendDolevKlaweRodehB(v, nil)
			} else {
				sendDolevKlaweRodehB(v, messageDolevKlaweRodehB{m2, m.Value, 0, 0})
			}
		}
	}

	data, _ := json.Marshal(s)
	v.SetState(data)

	return false
}

func runDolevKlaweRodehB(v lib.Node) {
	v.StartProcessing()
	finish := initializeDolevKlaweRodehB(v)
	v.FinishProcessing(finish)

	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = processDolevKlaweRodehB(v, round)
		v.FinishProcessing(finish)
	}
}

func checkDolevKlaweRodehB(vertices []lib.Node) {
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
		}
	}
}

func RunDolevKlaweRodehB(n int) (int, int) {
	vertices, synchronizer := lib.BuildSynchronizedDirectedRing(n)

	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runDolevKlaweRodehB(v)
	}
	synchronizer.Synchronize(0)
	checkDolevKlaweRodehB(vertices)

	return synchronizer.GetStats()
}
