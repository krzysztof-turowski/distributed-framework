package directed_ring

import (
	"encoding/json"
	"fmt"
	"lib"
	"log"
	"math/rand"
	"time"
)

type stateItaiRodeh struct {
	Counter int
	Active  int
	Token   bool
	Status  modeType
}

type messageItaiRodeh struct {
	Token bool
}

func saveStateItaiRodeh(v lib.Node, s stateItaiRodeh) {
	data, _ := json.Marshal(s)
	v.SetState(data)
}

func sendItaiRodeh(v lib.Node, s stateItaiRodeh, token bool) {
	if token {
		outMessage, _ := json.Marshal(messageItaiRodeh{Token: token})
		v.SendMessage(0, outMessage)
	} else {
		v.SendMessage(0, nil)
	}
	saveStateItaiRodeh(v, s)
}

func receiveItaiRodeh(v lib.Node) (stateItaiRodeh, messageItaiRodeh) {
	var s stateItaiRodeh
	json.Unmarshal(v.GetState(), &s)
	var m messageItaiRodeh
	inMessage := v.ReceiveMessage(0)
	json.Unmarshal(inMessage, &m)
	return s, m
}

func initializeItaiRodeh(v lib.Node) bool {
	rand.Seed(int64(v.GetIndex()))
	s := stateItaiRodeh{
		Counter: 0, Active: v.GetSize(),
		Status: unknown, Token: (rand.Intn(v.GetSize()) == 0)}
	sendItaiRodeh(v, s, s.Token)
	return false
}

func processItaiRodeh(v lib.Node, round int) bool {
	s, m := receiveItaiRodeh(v)
	if m.Token {
		s.Counter++
	}
	if round%v.GetSize() != 0 {
		sendItaiRodeh(v, s, m.Token)
	} else {
		if s.Status == unknown {
			if s.Counter == 1 {
				s.Active = 1
				if s.Token {
					s.Status = leader
				} else {
					s.Status = nonleader
				}
				saveStateItaiRodeh(v, s)
				return true
			} else {
				if s.Counter >= 2 {
					s.Active = s.Counter
					if !s.Token {
						s.Status = nonleader
					}
				}
				s.Counter, s.Token = 0, (rand.Intn(s.Active) == 0)
				sendItaiRodeh(v, s, s.Token && s.Status == unknown)
			}
		} else if s.Status == nonleader {
			if s.Counter == 1 {
				s.Active = 1
				saveStateItaiRodeh(v, s)
				return true
			} else if s.Counter >= 2 {
				s.Active = s.Counter
			}
			s.Counter = 0
			sendItaiRodeh(v, s, false)
		}
	}
	return false
}

func runItaiRodeh(v lib.Node) {
	v.StartProcessing()
	finish := initializeItaiRodeh(v)
	v.FinishProcessing(finish)

	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = processItaiRodeh(v, round)
		v.FinishProcessing(finish)
	}
}

func checkItaiRodeh(vertices []lib.Node) {
	var lead_node lib.Node
	var s stateItaiRodeh
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
		if v != lead_node {
			if s.Status == leader {
				panic(fmt.Sprintf(
					"Multiple leaders on the directed ring: %d, %d", lead_node.GetIndex(), v.GetIndex()))
			}
			if s.Status != nonleader {
				panic(fmt.Sprintf("Node %u has state %s", v.GetIndex(), s.Status))
			}
		}
		if s.Active != 1 {
			panic(fmt.Sprintf("Node %u has %d active nodes", v.GetIndex(), s.Active))
		}
	}
}

func RunItaiRodeh(n int) (int, int) {
	vertices, synchronizer := lib.BuildSynchronizedDirectedRing(n)
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runItaiRodeh(v)
	}
	synchronizer.Synchronize(0 * time.Millisecond)
	checkItaiRodeh(vertices)
	return synchronizer.GetStats()
}
