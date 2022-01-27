package undirected_ring

import (
	"encoding/json"
	"fmt"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
)

type modeType string

const (
	pass      modeType = "pass"
	unknown            = "unknown"
	nonleader          = "nonleader"
	leader             = "leader"
)

type messageType string

const (
	out = "out"
	in = "in"
	end = "end"
	null = "null"
)

type stateHirschbergSinclair struct {
	MyValue    int
	Status modeType
	MaxNum int
}

type messageHirschbergSinclair struct {
	MessageType messageType
	Value  int
	Num int
}

func sendHirschbergSinclair(v lib.Node, s stateHirschbergSinclair, mLeft messageHirschbergSinclair, mRight messageHirschbergSinclair) {
	data, _ := json.Marshal(s)
	v.SetState(data)
	if s.Status == leader {
		mEnd := messageHirschbergSinclair{MessageType: end}
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

func receiveHirschbergSinclair(v lib.Node) (stateHirschbergSinclair, messageHirschbergSinclair, messageHirschbergSinclair) {
	var s stateHirschbergSinclair
	json.Unmarshal(v.GetState(), &s)
	var mLeft, mRight messageHirschbergSinclair
	i, inMessage := v.ReceiveAnyMessage()
	if i==0{
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

func initializeHirschbergSinclair(v lib.Node) bool {
	s := stateHirschbergSinclair{MyValue: v.GetIndex(), Status: unknown, MaxNum: 1}
	sendHirschbergSinclair(v, s, messageHirschbergSinclair{
		MessageType: out,
		Value:       s.MyValue,
		Num:         s.MaxNum,
	}, messageHirschbergSinclair{
		MessageType: out,
		Value:       s.MyValue,
		Num:         s.MaxNum,
	})
	return false
}

func handleMessageHirschbergSinclair(s stateHirschbergSinclair, result bool, receivedA messageHirschbergSinclair,
	sendA messageHirschbergSinclair, sendB messageHirschbergSinclair) (messageHirschbergSinclair, messageHirschbergSinclair, stateHirschbergSinclair, bool){
	if receivedA.MessageType == out{
		if receivedA.Value > s.MyValue && receivedA.Num > 1 {
			sendB.MessageType = out
			sendB.Num = receivedA.Num - 1
			sendB.Value = receivedA.Value
		} else if receivedA.Value > s.MyValue && receivedA.Num == 1 {
			sendA.MessageType = in
			sendA.Num = 1
			sendA.Value = receivedA.Value
		} else if receivedA.Value == s.MyValue {
			s.Status = leader
			result = true
		}
	} else if receivedA.MessageType == in && receivedA.Value != s.MyValue{
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

func processHirschbergSinclair(v lib.Node, round int) bool {
	s, receivedLeft, receivedRight := receiveHirschbergSinclair(v)
	result := false
	sendLeft, sendRight := messageHirschbergSinclair{MessageType: null}, messageHirschbergSinclair{MessageType: null}
	sendLeft, sendRight, s, result = handleMessageHirschbergSinclair(s, result, receivedLeft, sendLeft, sendRight)
	sendRight, sendLeft, s, result = handleMessageHirschbergSinclair(s, result, receivedRight, sendRight, sendLeft)
	if receivedLeft.Value == s.MyValue && receivedLeft.MessageType == in && receivedLeft.Num == 1 &&
		receivedRight.Value == s.MyValue && receivedRight.MessageType == in && receivedRight.Num == 1 {
		s.MaxNum *= 2
		sendLeft.Value = s.MyValue
		sendLeft.MessageType = out
		sendLeft.Num = s.MaxNum
		sendRight.Value = s.MyValue
		sendRight.MessageType = out
		sendRight.Num = s.MaxNum
	}
	sendHirschbergSinclair(v, s, sendLeft, sendRight)
	return result
}

func runHirschbergSinclair(v lib.Node) {
	v.StartProcessing()
	finish := initializeHirschbergSinclair(v)
	v.FinishProcessing(finish)
	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = processHirschbergSinclair(v, round)
		v.FinishProcessing(finish)
	}
}

func checkHirschbergSinclair(vertices []lib.Node) {
	var leaderNode lib.Node
	var s stateHirschbergSinclair
	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if s.Status == leader {
			leaderNode = v
			break
		}
	}
	if leaderNode == nil {
		panic("There is no leader on the undirected ring")
	}
	max := 0
	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if s.MyValue > max{
			max = s.MyValue
		}
		if v != leaderNode {
			if s.Status == leader {
				panic(fmt.Sprint(
					"Multiple leaders on the undirected ring: ", s.MyValue, leaderNode.GetIndex()))
			}
			if s.Status != nonleader {
				panic(fmt.Sprint("Node ", v.GetIndex(), " has state ", s.Status))
			}
		}
	}
	if max > leaderNode.GetIndex(){
		panic(fmt.Sprint("Leader has value ", leaderNode.GetIndex(), " but max is ", max))
	}
}

func RunHirschbergSinclair(n int) (int, int){
	vertices, synchronizer := lib.BuildSynchronizedRing(n)
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runHirschbergSinclair(v)
	}
	synchronizer.Synchronize(0)
	checkHirschbergSinclair(vertices)

	return synchronizer.GetStats()
}
