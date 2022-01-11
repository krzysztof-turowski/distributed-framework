package clique

import (
	"log"
	"bytes"
	"encoding/binary"
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

const (
	msgCapture = iota
	msgAccept
	msgYes
	msgNo
	msgLeader
)

const none = -1;

type captureMessage struct {
	sender int
	passed bool
	level  int
	id     int
}

func RunHumblet(nodes []lib.Node, runner lib.Runner) int {
	for _, node := range nodes {
		log.Println("Running node", node.GetIndex())
		go runHumblet(node)
	}

	runner.Run()
	runner.GetStats()

	checkSingleLeaderElected(nodes)

	leader := getLeader(nodes[0])
	log.Println("Elected node", leader, "as a leader")

	return leader
}

func runHumblet(node lib.Node) {
	node.StartProcessing()

	id := node.GetIndex()
	size := node.GetSize()
	active := true
	level := 0
	owner := none
	queue := make([]captureMessage, 0)
	contender := none

	captureNext := func() {
		node.SendMessage(level, encodeAll(byte(msgCapture), int32(level), int64(id)))
	}

	passToOwner := func(message captureMessage) {
		node.SendMessage(owner, encodeAll(byte(msgCapture), int32(message.level), int64(message.id)))
	}

	respond := func(receiver int, label byte) {
		node.SendMessage(receiver, encodeAll(label))
	}

	announceAsLeader := func() {
		message := encodeAll(byte(msgLeader), int64(id))
		for receiver := 0; receiver < size - 1; receiver++ {
			node.SendMessage(receiver, message)
		}
	}

	log.Println("Node", id, "tries to capture first node")
	captureNext()

	loop: for {
		sender, message := node.ReceiveAnyMessage()
		buffer := bytes.NewBuffer(message)

		var label byte
		decode(buffer, &label)

		switch label {
		case msgCapture:
			passed := sender < level
			var level int32
			var id int64
			decode(buffer, &level, &id)
			queue = append(queue, captureMessage{sender: sender, passed: passed, level: int(level), id: int(id)})

		case msgAccept:
			level++
			if active {
				if level + 1 <= size / 2 {
					log.Println("Node", id, "reaches level", level, "and tries to capture next node")
					captureNext()
				} else {
					log.Println("Node", id, "reaches level", level, "and becomes a leader")
					setLeader(node, id)
					announceAsLeader()
					break loop
				}
			}

		case msgYes:
			log.Println("Node", id, "changes its owner")
			owner = contender
			respond(contender, msgAccept)
			contender = none

		case msgNo:
			log.Println("Node", id, "keeps its owner")
			contender = none

		case msgLeader:
			log.Println("Node", id, "receives leader announcement")
			var leader int64
			decode(buffer, &leader)
			setLeader(node, int(leader))
			break loop
		}

		for contender == none && len(queue) > 0 {
			message := queue[0]
			queue = queue[1:]

			compare := func() bool {
				return message.level > level || (message.level == level && message.id > id)
			}

			if message.passed {
				if compare() {
					log.Println("Node", id, "yields to node", message.id)
					active = false
					respond(message.sender, msgYes)
				} else {
					log.Println("Node", id, "ignores node", message.id)
					respond(message.sender, msgNo)
				}
			} else if owner == none {
				if compare() {
					log.Println("Node", id, "yields to node", message.id, "and becomes captured")
					active = false
					owner = message.sender
					respond(message.sender, msgAccept)
				} else {
					log.Println("Node", id, "ignores node", message.id)
				}
			} else {
				log.Println("Node", id, "passes a message from node", message.id, "to its owner")
				passToOwner(message)
				contender = message.sender
			}
		}
	}

	node.FinishProcessing(true)
}

func setLeader(node lib.Node, leader int) {
	value := int64(leader)
	node.SetState(encodeAll(value))
}

func getLeader(node lib.Node) int {
	var value int64
	decodeAll(node.GetState(), &value)
	return int(value)
}

func checkSingleLeaderElected(nodes []lib.Node) {
	leader := getLeader(nodes[0])
	for _, node := range nodes {
		if getLeader(node) != leader {
			panic("Multiple leaders elected")
		}
	}
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
