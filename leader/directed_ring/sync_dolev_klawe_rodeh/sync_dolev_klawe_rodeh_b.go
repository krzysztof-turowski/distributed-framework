package sync_dolev_klawe_rodeh

import (
	"encoding/json"
	"fmt"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
)

const (
	waiting status = 3
)

type stateB struct {
	PhaseV int
	MaxV   int
	Status status
}

const (
	empty messageType = 0 //util type
)

type messageB struct {
	MessageType messageType
	Value       int
	Phase       int
	Counter     int
}

func sendB(v lib.Node, message interface{}) {
	if message != nil {
		messageBytes, _ := json.Marshal(message)
		v.SendMessage(0, messageBytes)
	} else {
		v.SendMessage(0, nil)
	}
}

func receiveB(v lib.Node) (stateB, messageB) {
	var s stateB
	json.Unmarshal(v.GetState(), &s)

	var m messageB
	message := v.ReceiveMessage(0)
	json.Unmarshal(message, &m)

	return s, m
}

func initializeB(v lib.Node) bool {
	s := stateB{0, v.GetIndex(), active}

	data, _ := json.Marshal(s)
	v.SetState(data)

	sendB(v, messageB{1, s.MaxV, s.PhaseV, 1})
	return false
}

func processB(v lib.Node, round int) bool {
	s, m := receiveB(v)

	if m.MessageType == empty {
		if s.Status != maximum {
			sendB(v, nil)
		}
		return false
	}

	if m.Value == v.GetIndex() {
		s.Status = maximum
		data, _ := json.Marshal(s)
		v.SetState(data)
		sendB(v, messageB{terminate, 0, 0, 0})
		return false
	}

	if m.MessageType == terminate {
		if s.Status == maximum {
			return true
		}
		sendB(v, m)
		return true
	}

	if s.Status == active {
		if m.MessageType == m1 {
			if m.Value > s.MaxV {
				s.MaxV = m.Value
				s.Status = waiting
				sendB(v, nil) // nil
			} else {
				s.Status = passive
				sendB(v, messageB{m2, s.MaxV, 0, 0})
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
					sendB(v, messageB{m1, m.Value, m.Phase, m.Counter - 1})
				} else if m.Counter == 1 {
					s.Status = waiting
					sendB(v, messageB{m1, m.Value, m.Phase, 0})
					s.PhaseV = m.Phase
				}
			} else {
				sendB(v, messageB{m1, m.Value, m.Phase, 0})
			}
		} else if m.MessageType == m2 {
			s.PhaseV = s.PhaseV + 1
			s.Status = active
			sendB(v, messageB{m1, s.MaxV, s.PhaseV, 1})

		}
	} else if s.Status == passive {
		if m.MessageType == m1 {
			if m.Value >= s.MaxV && m.Counter >= 1 {
				s.MaxV = m.Value
				if m.Counter > 1 {
					sendB(v, messageB{m1, m.Value, m.Phase, m.Counter - 1})
				} else if m.Counter == 1 {
					s.Status = waiting
					sendB(v, messageB{m1, m.Value, m.Phase, 0})
					s.PhaseV = m.Phase
				}
			} else {
				sendB(v, messageB{m1, m.Value, m.Phase, 0})
			}
		} else if m.MessageType == 2 {
			if m.Value < s.MaxV {
				sendB(v, nil)
			} else {
				sendB(v, messageB{m2, m.Value, 0, 0})
			}
		}
	}

	data, _ := json.Marshal(s)
	v.SetState(data)

	return false
}

func runB(v lib.Node) {
	v.StartProcessing()
	finish := initializeB(v)
	v.FinishProcessing(finish)

	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = processB(v, round)
		v.FinishProcessing(finish)
	}
}

func checkB(vertices []lib.Node) {
	var leadNode lib.Node
	var s stateA

	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if s.Status == maximum {
			leadNode = v
			break
		}
	}

	if leadNode == nil {
		panic("There is no leader on the directed undirected_ring")
	}

	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if v != leadNode {
			if s.Status == maximum {
				panic(fmt.Sprint(
					"Multiple leaders on the directed undirected_ring: ", leadNode.GetIndex(), v.GetIndex()))
			}
		}
	}
}

func RunB(n int) (int, int) {
	vertices, synchronizer := lib.BuildSynchronizedDirectedRing(n)

	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runB(v)
	}
	synchronizer.Synchronize(0)
	checkB(vertices)

	return synchronizer.GetStats()
}
