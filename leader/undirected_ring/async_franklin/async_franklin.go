package async_franklin

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"log"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

const (
	LEFT  = 0
	RIGHT = 1
)

const (
	active = iota
	passive
	knownLeader
)

type state struct {
	Id       uint64
	IsLeader bool
	LeaderId uint64
	Phase    uint64
}

type messageContent struct {
	Value   uint64
	IsFinal bool
}

func run(node lib.Node) {
	node.StartProcessing()
	initialize(node)
	if process(node) {
		finishAsLeader(node)
	} else {
		//node.IgnoreFutureMessages()
	}
	node.FinishProcessing(true)
}

func process(node lib.Node) bool {
	for {
		switch getState(node).Phase {
		case active:
			handleActive(node)
		case passive:
			handlePassive(node)
		case knownLeader:
			return isLeader(node)
		}
	}
}

func initialize(node lib.Node) {
	id := uint64(node.GetIndex())
	setState(node, &state{
		Id:       id,
		IsLeader: false,
		LeaderId: 0,
		Phase:    active,
	})

	msg := messageContent{
		Value:   id,
		IsFinal: false,
	}

	node.SendMessage(LEFT, encodeAll(msg))
	node.SendMessage(RIGHT, encodeAll(msg))
}

func handleActive(node lib.Node) {
	var valLeft uint64
	var valRight uint64
	var gotFinal bool
	sender, byteMsg := node.ReceiveAnyMessage()

	if sender == LEFT {
		valLeft, valRight, gotFinal = handleTwoMessages(node, LEFT, RIGHT, byteMsg)
	} else {
		valRight, valLeft, gotFinal = handleTwoMessages(node, RIGHT, LEFT, byteMsg)
	}
	if gotFinal {
		return
	}
	id := getState(node).Id
	if valLeft > id || valRight > id {
		setPassiveState(node)
	} else if valLeft < id && valRight < id {
		byteMsg = encodeAll(messageContent{
			Value:   id,
			IsFinal: false,
		})
		node.SendMessage(LEFT, byteMsg)
		node.SendMessage(RIGHT, byteMsg)
	} else {
		log.Println("[LEADER] Node", node.GetIndex(), "becomes leader")
		setFinalState(node, id, true)
	}
}

func handleTwoMessages(node lib.Node, first int, second int, firstMsgBytes []byte) (uint64, uint64, bool) {
	var firstMsg messageContent
	var secondMsg messageContent
	byteMsg := firstMsgBytes

	decodeAll(byteMsg, &firstMsg)
	if firstMsg.IsFinal {
		log.Println("Node ", node.GetIndex(), " got final message")
		node.SendMessage(second, byteMsg)
		setFinalState(node, firstMsg.Value, false)
		return 0, 0, true
	}
	byteMsg = node.ReceiveMessage(second)
	decodeAll(byteMsg, &secondMsg)
	if secondMsg.IsFinal {
		log.Println("Node ", node.GetIndex(), " got final message")
		node.SendMessage(first, byteMsg)
		setFinalState(node, secondMsg.Value, false)
		return 0, 0, true
	}
	return firstMsg.Value, secondMsg.Value, false
}

func handlePassive(node lib.Node) {
	var sender int
	var byteMsg []byte
	var msg messageContent

	log.Println("Node", node.GetIndex(), "becomes passive")
	for {
		sender, byteMsg = node.ReceiveAnyMessage()
		if sender == LEFT {
			decodeAll(byteMsg, &msg)
			node.SendMessage(RIGHT, byteMsg)
			if msg.IsFinal {
				setFinalState(node, msg.Value, false)
				return
			}
		} else {
			decodeAll(byteMsg, &msg)
			node.SendMessage(LEFT, byteMsg)
			if msg.IsFinal {
				setFinalState(node, msg.Value, false)
				return
			}
		}
	}
}

func setFinalState(node lib.Node, leaderId uint64, isLeader bool) {
	setState(node, &state{
		Id:       getState(node).Id,
		IsLeader: isLeader,
		LeaderId: leaderId,
		Phase:    knownLeader,
	})
}

func setPassiveState(node lib.Node) {
	setState(node, &state{
		Id:       getState(node).Id,
		IsLeader: false,
		LeaderId: 0,
		Phase:    passive,
	})
}

func finishAsLeader(node lib.Node) {
	var msg messageContent

	log.Println("Leader node", getState(node).Id, "sends final message")
	node.SendMessage(LEFT, encodeAll(messageContent{
		Value:   getState(node).Id,
		IsFinal: true,
	}))

	for {
		sender, byteMsg := node.ReceiveAnyMessage()
		decodeAll(byteMsg, &msg)
		if sender == RIGHT && msg.IsFinal {
			return
		}
	}
}

//main algorithm running

func Run(nodes []lib.Node, runner lib.Runner) (int, int) {
	for _, node := range nodes {
		log.Println("Running node", node.GetIndex())
		go run(node)
	}
	runner.Run(true)
	checkSingleLeaderElected(nodes)

	return runner.GetStats()
}

func checkSingleLeaderElected(nodes []lib.Node) {
	leader := getLeader(nodes[0])
	for _, node := range nodes {
		if getLeader(node) != leader {
			panic("Multiple leaders elected")
		}
	}
	log.Println("Elected node", leader, "as a leader")
}

//state handling

func getState(node lib.Node) state {
	var state state
	json.Unmarshal(node.GetState(), &state)
	return state
}

func setState(node lib.Node, state *state) {
	encodedState, _ := json.Marshal(*state)
	node.SetState(encodedState)
}

func isLeader(node lib.Node) bool {
	return isLeaderKnown(node) && getLeader(node) == getState(node).Id
}

func getLeader(node lib.Node) uint64 {
	if !isLeaderKnown(node) {
		panic("Leader value is not known")
	}
	return getState(node).LeaderId
}

func isLeaderKnown(node lib.Node) bool {
	return getState(node).Phase == knownLeader
}

//encoding and decoding

func encodeAll(values ...interface{}) []byte {
	buffer := bytes.Buffer{}
	encode(&buffer, values...)
	return buffer.Bytes()
}

func decodeAll(data []byte, values ...interface{}) {
	buffer := bytes.NewBuffer(data)
	decode(buffer, values...)
}

func encode(buffer *bytes.Buffer, values ...interface{}) {
	for _, value := range values {
		if binary.Write(buffer, binary.BigEndian, value) != nil {
			panic("Failed to encode a value")
		}
	}
}

func decode(buffer *bytes.Buffer, values ...interface{}) {
	for _, value := range values {
		if binary.Read(buffer, binary.BigEndian, value) != nil {
			panic("Failed to decode a value")
		}
	}
}
