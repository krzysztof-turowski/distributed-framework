package sync_itai_rodeh

import (
	"encoding/json"
	"fmt"
	. "github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/common"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
	"math/rand"
	"time"
)

type state struct {
	Counter int
	Active  int
	Token   bool
	Status  ModeType
}

type message struct {
	Token bool
}

func saveState(v lib.Node, s state) {
	data, _ := json.Marshal(s)
	v.SetState(data)
}

func send(v lib.Node, s state, token bool) {
	if token {
		outMessage, _ := json.Marshal(message{Token: token})
		v.SendMessage(0, outMessage)
	} else {
		v.SendMessage(0, nil)
	}
	saveState(v, s)
}

func receive(v lib.Node) (state, message) {
	var s state
	json.Unmarshal(v.GetState(), &s)
	var m message
	inMessage := v.ReceiveMessage(0)
	json.Unmarshal(inMessage, &m)
	return s, m
}

func initialize(v lib.Node) bool {
	rand.Seed(int64(v.GetIndex()))
	s := state{
		Counter: 0, Active: v.GetSize(),
		Status: Unknown, Token: (rand.Intn(v.GetSize()) == 0)}
	send(v, s, s.Token)
	return false
}

func process(v lib.Node, round int) bool {
	s, m := receive(v)
	if m.Token {
		s.Counter++
	}
	if round%v.GetSize() != 0 {
		send(v, s, m.Token)
	} else {
		if s.Status == Unknown {
			if s.Counter == 1 {
				s.Active = 1
				if s.Token {
					s.Status = Leader
				} else {
					s.Status = NonLeader
				}
				saveState(v, s)
				return true
			} else {
				if s.Counter >= 2 {
					s.Active = s.Counter
					if !s.Token {
						s.Status = NonLeader
					}
				}
				s.Counter, s.Token = 0, (rand.Intn(s.Active) == 0)
				send(v, s, s.Token && s.Status == Unknown)
			}
		} else if s.Status == NonLeader {
			if s.Counter == 1 {
				s.Active = 1
				saveState(v, s)
				return true
			} else if s.Counter >= 2 {
				s.Active = s.Counter
			}
			s.Counter = 0
			send(v, s, false)
		}
	}
	return false
}

func run(v lib.Node) {
	v.StartProcessing()
	finish := initialize(v)
	v.FinishProcessing(finish)

	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = process(v, round)
		v.FinishProcessing(finish)
	}
}

func check(vertices []lib.Node) {
	var leadNode lib.Node
	var s state
	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if s.Status == Leader {
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
			if s.Status == Leader {
				panic(fmt.Sprintf(
					"Multiple leaders on the directed undirected_ring: %d, %d", leadNode.GetIndex(), v.GetIndex()))
			}
			if s.Status != NonLeader {
				panic(fmt.Sprintf("Node %d has state %s", v.GetIndex(), s.Status))
			}
		}
		if s.Active != 1 {
			panic(fmt.Sprintf("Node %d has %d active nodes", v.GetIndex(), s.Active))
		}
	}
}

func Run(n int) (int, int) {
	vertices, synchronizer := lib.BuildSynchronizedDirectedRing(n)
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go run(v)
	}
	synchronizer.Synchronize(0 * time.Millisecond)
	check(vertices)
	return synchronizer.GetStats()
}
