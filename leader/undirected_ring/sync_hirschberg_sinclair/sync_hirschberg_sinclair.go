package sync_hirschberg_sinclair

import (
	"encoding/json"
	"fmt"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
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
	Status modeType
	MaxNum int
}

type message struct {
	MessageType messageType
	Value       int
	Num         int
}

func send(v lib.Node, s state, mLeft message, mRight message) {
	data, _ := json.Marshal(s)
	v.SetState(data)
	if s.Status == leader {
		mEnd := message{MessageType: end}
		endMessage, _ := json.Marshal(mEnd)
		v.SendMessage(0, endMessage)
		v.SendMessage(1, endMessage)
		return
	}
	if mLeft.MessageType == end && mRight.MessageType == end {
		return
	}
	if s.Status != nonleader || (s.Status == nonleader && mLeft.MessageType == end) {
		if mLeft.MessageType != null {
			leftMessage, _ := json.Marshal(mLeft)
			v.SendMessage(0, leftMessage)
		} else {
			v.SendMessage(0, nil)
		}
	}
	if s.Status != nonleader || (s.Status == nonleader && mRight.MessageType == end) {
		if (s.Status != nonleader && mRight.MessageType != null) || (s.Status == nonleader && mRight.MessageType == end) {
			rightMessage, _ := json.Marshal(mRight)
			v.SendMessage(1, rightMessage)
		} else {
			v.SendMessage(1, nil)
		}
	}

}

func receive(v lib.Node) (state, message, message) {
	var s state
	json.Unmarshal(v.GetState(), &s)
	var mLeft, mRight message
	i, inMessage := v.ReceiveAnyMessage()
	if i == 0 {
		json.Unmarshal(inMessage, &mLeft)
		inMessage = v.ReceiveMessage(1)
		json.Unmarshal(inMessage, &mRight)
	} else {
		json.Unmarshal(inMessage, &mRight)
		inMessage = v.ReceiveMessage(0)
		json.Unmarshal(inMessage, &mLeft)
	}
	return s, mLeft, mRight
}

func initialize(v lib.Node) bool {
	s := state{Status: unknown, MaxNum: 1}
	log.Println("Node", v.GetIndex(), "initiates a message with length", s.MaxNum)
	send(v, s, message{
		MessageType: out,
		Value:       v.GetIndex(),
		Num:         s.MaxNum,
	}, message{
		MessageType: out,
		Value:       v.GetIndex(),
		Num:         s.MaxNum,
	})
	return false
}

func handleMessage(v lib.Node, s state, result bool, receivedA message,
	sendA message, sendB message) (message, message, state, bool) {
	if receivedA.MessageType == out {
		if receivedA.Value > v.GetIndex() && receivedA.Num > 1 {
			sendB.MessageType = out
			sendB.Num = receivedA.Num - 1
			sendB.Value = receivedA.Value
		} else if receivedA.Value > v.GetIndex() && receivedA.Num == 1 {
			sendA.MessageType = in
			sendA.Num = 1
			sendA.Value = receivedA.Value
		} else if receivedA.Value == v.GetIndex() {
			s.Status = leader
			result = true
		}
	} else if receivedA.MessageType == in && receivedA.Value != v.GetIndex() {
		sendB.MessageType = in
		sendB.Num = 1
		sendB.Value = receivedA.Value
	} else if receivedA.MessageType == end {
		if s.Status != leader {
			sendB.MessageType = end
			s.Status = nonleader
		}
		result = true
	}
	return sendA, sendB, s, result
}

func process(v lib.Node, round int) bool {
	s, receivedLeft, receivedRight := receive(v)
	result := false
	sendLeft, sendRight := message{MessageType: null}, message{MessageType: null}
	sendLeft, sendRight, s, result = handleMessage(v, s, result, receivedLeft, sendLeft, sendRight)
	sendRight, sendLeft, s, result = handleMessage(v, s, result, receivedRight, sendRight, sendLeft)
	if receivedLeft.Value == v.GetIndex() && receivedLeft.MessageType == in && receivedLeft.Num == 1 &&
		receivedRight.Value == v.GetIndex() && receivedRight.MessageType == in && receivedRight.Num == 1 {
		s.MaxNum *= 2
		sendLeft.Value = v.GetIndex()
		sendLeft.MessageType = out
		sendLeft.Num = s.MaxNum
		sendRight.Value = v.GetIndex()
		sendRight.MessageType = out
		sendRight.Num = s.MaxNum
		log.Println("Node", v.GetIndex(), "initiates a message with length", s.MaxNum)
	}
	send(v, s, sendLeft, sendRight)
	return result
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
	vertices, synchronizer := lib.BuildSynchronizedRing(n)
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go run(v)
	}
	synchronizer.Synchronize(0)
	check(vertices)

	return synchronizer.GetStats()
}
