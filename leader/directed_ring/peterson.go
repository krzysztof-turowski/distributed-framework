package directed_ring

import (
	"encoding/json"
	"fmt"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
)

type statePeterson struct {
	Max    int
	Status modeType
}

type messagePeterson struct {
	Max  int
	Mode modeType
}

func initializePeterson(v lib.Node) bool {
	s := statePeterson{Max: v.GetIndex(), Status: unknown}
	data, _ := json.Marshal(s)
	v.SetState(data)
	return false
}

func sendMessagePeterson(v lib.Node, m messagePeterson) {
	outMessage, _ := json.Marshal(m)
	v.SendMessage(0, outMessage)
}

func receiveMessagePeterson(v lib.Node) messagePeterson {
	var m messagePeterson
	inMessage := v.ReceiveMessage(0)
	json.Unmarshal(inMessage, &m)
	return m
}

func getStatePeterson(v lib.Node) statePeterson {
	var s statePeterson
	json.Unmarshal(v.GetState(), &s)
	return s
}

func setStatePeterson(v lib.Node, s statePeterson) {
	data, _ := json.Marshal(s)
	v.SetState(data)
}

func processPeterson(v lib.Node, round int) bool {
	s := getStatePeterson(v)
	if s.Status != relay {
		if s.Status == unknown {
			sendMessagePeterson(v, messagePeterson{Max: s.Max, Mode: unknown})
			m1 := receiveMessagePeterson(v)
			if m1.Mode == leader {
				s.Max = m1.Max
				s.Status = nonleader
				sendMessagePeterson(v, messagePeterson{Max: m1.Max, Mode: leader})
			} else if m1.Max == v.GetIndex() {
				s.Status = leader
				sendMessagePeterson(v, messagePeterson{Max: s.Max, Mode: leader})
			} else {
				if s.Max > m1.Max {
					sendMessagePeterson(v, messagePeterson{Max: s.Max, Mode: unknown})
				} else {
					sendMessagePeterson(v, messagePeterson{Max: m1.Max, Mode: unknown})
				}
				m2 := receiveMessagePeterson(v)
				if m2.Mode == leader {
					s.Max = m2.Max
					s.Status = nonleader
					sendMessagePeterson(v, messagePeterson{Max: m2.Max, Mode: leader})
				} else if m2.Max == v.GetIndex() {
					s.Status = leader
					sendMessagePeterson(v, messagePeterson{Max: s.Max, Mode: leader})
				} else if m1.Max >= s.Max && m1.Max >= m2.Max {
					s.Max = m1.Max
				} else {
					s.Status = relay
				}
			}
			setStatePeterson(v, s)
			if s.Status == nonleader {
				return true
			}
		} else if s.Status == leader {
			m := receiveMessagePeterson(v)
			if m.Mode == leader {
				return true
			}
		}
	} else {
		for i := 0; i < 2; i++ {
			m := receiveMessagePeterson(v)
			if m.Max > s.Max {
				s.Max = m.Max
			}
			if m.Mode == leader {
				if s.Status != leader {
					s.Status = nonleader
				}
				setStatePeterson(v, s)
				sendMessagePeterson(v, m)
				return true
			} else if m.Max == v.GetIndex() {
				s.Status = leader
				sendMessagePeterson(v, messagePeterson{Max: m.Max, Mode: leader})
			} else {
				sendMessagePeterson(v, m)
			}
			setStatePeterson(v, s)
		}
	}
	return false
}

func runPeterson(v lib.Node) {
	v.StartProcessing()
	finish := initializePeterson(v)
	v.FinishProcessing(finish)

	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = processPeterson(v, round)
		v.FinishProcessing(finish)
	}
}

func checkPeterson(vertices []lib.Node) {
	var lead_node lib.Node
	var s statePeterson
	for _, v := range vertices {
		s = getStatePeterson(v)
		if s.Status == leader {
			lead_node = v
			break
		}
	}
	if lead_node == nil {
		panic("There is no leader on the directed ring")
	}
	for _, v := range vertices {
		s = getStatePeterson(v)
		if v == lead_node {
			if v.GetIndex() != s.Max {
				panic(fmt.Sprint("Leader has index ", v.GetIndex(), " but maximum ", s.Max))
			}
		} else {
			if s.Status == leader {
				panic(fmt.Sprint(
					"Multiple leaders on the directed ring: ", lead_node.GetIndex(), v.GetIndex()))
			}
			if s.Status != nonleader {
				panic(fmt.Sprint("Node ", v.GetIndex(), " has state ", s.Status))
			}
			if lead_node.GetIndex() != s.Max {
				panic(fmt.Sprint(
					"Leader has index ", lead_node.GetIndex(), " but node ", v.GetIndex(),
					" has maximum ", s.Max))
			}
		}
	}
}

func RunPeterson(n int) (int, int) {
	vertices, synchronizer := lib.BuildSynchronizedDirectedRing(n)
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runPeterson(v)
	}
	synchronizer.Synchronize(0)
	checkPeterson(vertices)

	return synchronizer.GetStats()
}
