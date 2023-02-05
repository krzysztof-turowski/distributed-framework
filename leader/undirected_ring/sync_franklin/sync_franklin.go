package sync_franklin

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

type modeType string

const (
	unknown   modeType = "unknown"
	nonLeader modeType = "nonleader"
	leader    modeType = "leader"
)

type messageType int

const (
	null messageType = iota
	normal
	ending
	empty
)

type nodeState struct {
	Status modeType
}

type message struct {
	MessageType messageType
	Value       int
}

const (
	indLeft  = 0
	indRight = 1
)

func encodeState(s nodeState) []byte {
	data, _ := json.Marshal(s)
	return data
}

func decodeState(encState []byte) nodeState {
	var result nodeState
	json.Unmarshal(encState, &result)
	return result
}

func decodeMessage(encMsg []byte) message {
	var result message
	if encMsg == nil {
		result = message{MessageType: empty}
	} else {
		json.Unmarshal(encMsg, &result)
	}
	return result
}

func sendBothMessages(v lib.Node, leftMsg message, rightMsg message) {
	if leftMsg.MessageType == empty {
		v.SendMessage(indLeft, nil)
	} else if leftMsg.MessageType != null {
		encLeftMsg, _ := json.Marshal(leftMsg)
		v.SendMessage(indLeft, encLeftMsg)
	}

	if rightMsg.MessageType == empty {
		v.SendMessage(indRight, nil)
	} else if rightMsg.MessageType != null {
		encRightMsg, _ := json.Marshal(rightMsg)
		v.SendMessage(indRight, encRightMsg)
	}
}

func receiveBothMessages(v lib.Node) (message, message) {
	var leftMsg, rightMsg message
	i, recEncMsg := v.ReceiveAnyMessage()
	if i == indLeft {
		leftMsg = decodeMessage(recEncMsg)
		recEncMsg = v.ReceiveMessage(indRight)
		rightMsg = decodeMessage(recEncMsg)
	} else {
		rightMsg = decodeMessage(recEncMsg)
		recEncMsg = v.ReceiveMessage(indLeft)
		leftMsg = decodeMessage(recEncMsg)
	}
	return leftMsg, rightMsg
}

func initialize(v lib.Node) bool {
	v.SetState(encodeState(nodeState{Status: unknown}))
	sendBothMessages(v,
		message{MessageType: normal, Value: v.GetIndex()},
		message{MessageType: normal, Value: v.GetIndex()})
	return false
}

func processActive(v lib.Node) bool {
	recLeft, recRight := receiveBothMessages(v)
	if (recLeft.MessageType == normal && recLeft.Value > v.GetIndex()) ||
		(recRight.MessageType == normal && recRight.Value > v.GetIndex()) {
		v.SetState(encodeState(nodeState{Status: nonLeader}))
		sendBothMessages(v,
			message{MessageType: empty},
			message{MessageType: empty})
		return false
	} else if (recLeft.MessageType == normal && recLeft.Value == v.GetIndex()) ||
		(recRight.MessageType == normal && recRight.Value == v.GetIndex()) {
		v.SetState(encodeState(nodeState{Status: leader}))
		sendBothMessages(v,
			message{MessageType: ending},
			message{MessageType: ending})
		return true
	} else {
		var toSendLeft, toSendRight message
		if recLeft.MessageType == empty {
			toSendLeft = message{MessageType: empty}
		} else {
			toSendLeft = message{MessageType: normal, Value: v.GetIndex()}
		}
		if recRight.MessageType == empty {
			toSendRight = message{MessageType: empty}
		} else {
			toSendRight = message{MessageType: normal, Value: v.GetIndex()}
		}
		sendBothMessages(v,
			toSendLeft,
			toSendRight)
		return false
	}
}

func processPassive(v lib.Node) bool {
	recLeft, recRight := receiveBothMessages(v)
	if recLeft.MessageType != ending && recRight.MessageType != ending {
		sendBothMessages(v, recRight, recLeft)
		return false
	} else if recLeft.MessageType != ending {
		sendBothMessages(v, recRight, message{MessageType: null})
		return true
	} else if recRight.MessageType != ending {
		sendBothMessages(v, message{MessageType: null}, recLeft)
		return true
	} else {
		return true
	}
}

func process(v lib.Node) bool {
	vState := decodeState(v.GetState())
	if vState.Status == unknown {
		return processActive(v)
	} else {
		return processPassive(v)
	}
}

func run(v lib.Node) {
	v.StartProcessing()
	isFinished := initialize(v)
	v.FinishProcessing(isFinished)
	for !isFinished {
		v.StartProcessing()
		isFinished = process(v)
		v.FinishProcessing(isFinished)
	}
}

func check(vertices []lib.Node) {
	var leadNode lib.Node
	var s nodeState
	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if s.Status == leader {
			leadNode = v
			break
		}
	}
	if leadNode == nil {
		panic("There is no leader on the undirected ring")
	}
	max := 0
	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if v.GetIndex() > max {
			max = v.GetIndex()
		}
		if v != leadNode {
			if s.Status == leader {
				panic(fmt.Sprint("Multiple leaders on the undirected ring: ",
					v.GetIndex(), leadNode.GetIndex()))
			}
			if s.Status != nonLeader {
				panic(fmt.Sprint("Node ", v.GetIndex(), " has state ", s.Status))
			}
		}
	}
	if max != leadNode.GetIndex() {
		panic(fmt.Sprint("Leader has value ", leadNode.GetIndex(), " but max is ", max))
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
