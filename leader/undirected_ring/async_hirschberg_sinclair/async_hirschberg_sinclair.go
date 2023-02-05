package async_hirschberg_sinclair

import (
	"encoding/json"
	"fmt"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
	"sync"
)

type modeType string

const (
	unknown   modeType = "unknown"
	nonleader          = "nonleader"
	leader             = "leader"
)

type messageType int

const (
	null messageType = iota
	out
	in
	end
)

type state struct {
	Status             modeType
	MaxNum             int
	ComebackCounter    int
	LastMessageCounter int
}

type message struct {
	MessageType messageType
	Value       int
	Num         int
}

func send(v lib.Node, m message, isLeft bool, wg *sync.WaitGroup) {
	var sendIndex int
	if isLeft {
		sendIndex = 0
	} else {
		sendIndex = 1
	}

	message, _ := json.Marshal(m)

	wg.Add(1)
	go func() {
		defer wg.Done()
		v.SendMessage(sendIndex, message)
	}()
}

func receive(v lib.Node) (message, bool) {
	var m message
	var isLeft bool
	i, inMessage := v.ReceiveAnyMessage()
	json.Unmarshal(inMessage, &m)
	if i == 0 {
		isLeft = true
	} else {
		isLeft = false
	}
	return m, isLeft
}

func initialize(v lib.Node, wg *sync.WaitGroup) bool {
	s := state{Status: unknown, MaxNum: 1, ComebackCounter: 0}
	data, _ := json.Marshal(s)
	v.SetState(data)
	log.Println("Node", v.GetIndex(), "initiates a message with length", s.MaxNum)
	send(v, message{
		MessageType: out,
		Value:       v.GetIndex(),
		Num:         s.MaxNum,
	}, true, wg)
	send(v, message{
		MessageType: out,
		Value:       v.GetIndex(),
		Num:         s.MaxNum,
	}, false, wg)
	return false
}

func handleMessage(v lib.Node, s state, result bool, receivedA message,
	sendA message, sendB message) (message, message, state, bool) {
	if receivedA.MessageType == out {
		if s.LastMessageCounter > 0 {
			return sendA, sendB, s, result
		} else if receivedA.Value == v.GetIndex() {
			s.ComebackCounter++
			if s.ComebackCounter == 1 {
				return sendA, sendB, s, result
			}
			sendA.MessageType = end
			sendA.Value = v.GetIndex()
			sendB.MessageType = end
			sendB.Value = v.GetIndex()
		} else if receivedA.Value > v.GetIndex() {
			receivedA.Num--
			if receivedA.Num > 0 {
				sendB.MessageType = out
				sendB.Value = receivedA.Value
				sendB.Num = receivedA.Num
			} else {
				sendA.MessageType = in
				sendA.Value = receivedA.Value
			}
		}
	} else if receivedA.MessageType == in {
		if s.LastMessageCounter > 0 {
			return sendA, sendB, s, result
		} else if receivedA.Value == v.GetIndex() {
			s.ComebackCounter++
			if s.ComebackCounter == 2 {
				s.ComebackCounter = 0
				s.MaxNum *= 2
				sendB.MessageType = out
				sendB.Value = v.GetIndex()
				sendB.Num = s.MaxNum
				sendA.MessageType = out
				sendA.Value = v.GetIndex()
				sendA.Num = s.MaxNum
				log.Println("Node", v.GetIndex(), "initiates a message with length", s.MaxNum)
			}
		} else {
			sendB.MessageType = in
			sendB.Value = receivedA.Value
		}
	} else if receivedA.MessageType == end {
		s.LastMessageCounter++
		if receivedA.Value != v.GetIndex() {
			s.Status = nonleader
			sendB.MessageType = end
			sendB.Value = receivedA.Value
		} else {
			s.Status = leader
		}
		if s.LastMessageCounter == 2 {
			result = true
		}
	}
	return sendA, sendB, s, result
}

func process(v lib.Node, wg *sync.WaitGroup) bool {
	var st state
	json.Unmarshal(v.GetState(), &st)
	received, isLeft := receive(v)
	result := false
	sendLeft, sendRight := message{MessageType: null}, message{MessageType: null}
	if isLeft {
		sendLeft, sendRight, st, result = handleMessage(v, st, result, received, sendLeft, sendRight)
	} else {
		sendRight, sendLeft, st, result = handleMessage(v, st, result, received, sendRight, sendLeft)
	}
	data, _ := json.Marshal(st)
	v.SetState(data)

	if sendLeft.MessageType != null {
		send(v, sendLeft, true, wg)
	}
	if sendRight.MessageType != null {
		send(v, sendRight, false, wg)
	}
	return result
}

func run(v lib.Node) {
	var wg sync.WaitGroup
	v.StartProcessing()
	initialize(v, &wg)
	for !process(v, &wg) {
	}
	wg.Wait()
	v.FinishProcessing(true)
}

func check(vertices []lib.Node) {
	var leaderNode lib.Node
	var s state
	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if s.Status == leader {
			leaderNode = v
			break
		}
	}
	if leaderNode == nil {
		panic("There is no leader on the undirected undirected_ring")
	}
	max := 0
	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if v.GetIndex() > max {
			max = v.GetIndex()
		}
		if v != leaderNode {
			if s.Status == leader {
				panic(fmt.Sprint(
					"Multiple leaders on the undirected undirected_ring: ", v.GetIndex(), leaderNode.GetIndex()))
			}
			if s.Status != nonleader {
				panic(fmt.Sprint("Node ", v.GetIndex(), " has state ", s.Status))
			}
		}
	}
	if max != leaderNode.GetIndex() {
		panic(fmt.Sprint("Leader has value ", leaderNode.GetIndex(), " but max is ", max))
	}
}

func Run(n int) (int, int) {
	vertices, runner := lib.BuildRing(n)
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go run(v)
	}
	runner.Run(true)
	check(vertices)

	return runner.GetStats()
}
