package clique

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
)

const (
	msgCapture = iota
	msgAccept
	msgYes
	msgNo
	msgLeader
)

const none = -1

type stateHumblet struct {
	Active    bool
	Level     int
	Owner     int
	Queue     []captureMessage
	Contender int
	Leader    int
}

type captureMessage struct {
	Sender int
	Passed bool
	Level  int
	Id     int
}

func RunHumblet(nodes []lib.Node, runner lib.Runner) int {
	for _, node := range nodes {
		log.Println("Running node", node.GetIndex())
		go runHumblet(node)
	}

	runner.Run()
	checkSingleLeaderElected(nodes)

	return runner.GetStats()
}

func runHumblet(node lib.Node) {
	node.StartProcessing()
	initializeHumblet(node)

	for {
		sender, message := node.ReceiveAnyMessage()
		if processHumblet(node, sender, message) {
			break
		}
	}
	node.FinishProcessing(true)
}

func initializeHumblet(node lib.Node) {
	setState(node, &stateHumblet{
		Active:    true,
		Level:     0,
		Owner:     none,
		Queue:     make([]captureMessage, 0),
		Contender: none,
		Leader:    none,
	})

	id := node.GetIndex()
	log.Println("Node", id, "tries to capture first node")
	node.SendMessage(0, encodeAll(byte(msgCapture), int32(0), int64(id)))
}

func processHumblet(node lib.Node, sender int, message []byte) bool {
	id := node.GetIndex()
	size := node.GetSize()
	state := getState(node)

	defer setState(node, &state)

	captureNext := func() {
		node.SendMessage(state.Level, encodeAll(byte(msgCapture), int32(state.Level), int64(id)))
	}

	passToOwner := func(message captureMessage) {
		node.SendMessage(state.Owner, encodeAll(byte(msgCapture), int32(message.Level), int64(message.Id)))
	}

	respond := func(receiver int, label byte) {
		node.SendMessage(receiver, encodeAll(label))
	}

	announceAsLeader := func() {
		message := encodeAll(byte(msgLeader), int64(id))
		for receiver := 0; receiver < size-1; receiver++ {
			node.SendMessage(receiver, message)
		}
	}

	buffer := bytes.NewBuffer(message)

	var label byte
	decode(buffer, &label)

	switch label {
	case msgCapture:
		passed := sender < state.Level
		var level int32
		var id int64
		decode(buffer, &level, &id)
		state.Queue = append(state.Queue, captureMessage{Sender: sender, Passed: passed, Level: int(level), Id: int(id)})

	case msgAccept:
		state.Level++
		if state.Active {
			if state.Level+1 <= size/2 {
				log.Println("Node", id, "reaches level", state.Level, "and tries to capture next node")
				captureNext()
			} else {
				log.Println("Node", id, "reaches level", state.Level, "and becomes a leader")
				state.Leader = id
				announceAsLeader()
				return true
			}
		}

	case msgYes:
		log.Println("Node", id, "changes its owner")
		state.Owner = state.Contender
		respond(state.Contender, msgAccept)
		state.Contender = none

	case msgNo:
		log.Println("Node", id, "keeps its owner")
		state.Contender = none

	case msgLeader:
		log.Println("Node", id, "receives leader announcement")
		var leader int64
		decode(buffer, &leader)
		state.Leader = int(leader)
		return true
	}

	for state.Contender == none && len(state.Queue) > 0 {
		message := state.Queue[0]
		state.Queue = state.Queue[1:]

		compare := func() bool {
			return message.Level > state.Level || (message.Level == state.Level && message.Id > id)
		}

		if message.Passed {
			if compare() {
				log.Println("Node", id, "yields to node", message.Id)
				state.Active = false
				respond(message.Sender, msgYes)
			} else {
				log.Println("Node", id, "ignores node", message.Id)
				respond(message.Sender, msgNo)
			}
		} else if state.Owner == none {
			if compare() {
				log.Println("Node", id, "yields to node", message.Id, "and becomes captured")
				state.Active = false
				state.Owner = message.Sender
				respond(message.Sender, msgAccept)
			} else {
				log.Println("Node", id, "ignores node", message.Id)
			}
		} else {
			log.Println("Node", id, "passes a message from node", message.Id, "to its owner")
			passToOwner(message)
			state.Contender = message.Sender
		}
	}

	return false
}

func getState(node lib.Node) stateHumblet {
	var state stateHumblet
	json.Unmarshal(node.GetState(), &state)
	return state
}

func setState(node lib.Node, state *stateHumblet) {
	encodedState, _ := json.Marshal(*state)
	node.SetState(encodedState)
}

func checkSingleLeaderElected(nodes []lib.Node) {
	leader := getState(nodes[0]).Leader
	for _, node := range nodes {
		if getState(node).Leader != leader {
			panic("Multiple leaders elected")
		}
	}
	log.Println("Elected node", leader, "as a leader")
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

func encodeAll(values ...interface{}) []byte {
	buffer := bytes.Buffer{}
	encode(&buffer, values...)
	return buffer.Bytes()
}

func decodeAll(data []byte, values ...interface{}) {
	buffer := bytes.NewBuffer(data)
	decode(buffer, values...)
}
