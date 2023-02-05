package async_probabilistic_franklin

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

type nodeStatus string

const (
	active    nodeStatus = "active"
	passive              = "passive"
	leader               = "leader"
	nonleader            = "nonleader"
)

type message struct {
	Id  int
	Hop int
	Bit int
}

type storedMessages struct {
	LeftMessages  [2]message
	RightMessages [2]message
}

type state struct {
	Id         int
	Status     nodeStatus
	DomainSize int
	Messages   storedMessages
	Bit        int
}

func newEmptyMessage() message {
	return message{-1, 0, 0}
}

func isMessageEmpty(m message) bool {
	return (m.Id == -1)
}

func newLeaderMessage() message {
	return message{-2, 0, 0}
}

func isMessageFromLeader(m message) bool {
	return (m.Id == -2)
}

func setState(node lib.Node, s state) {
	encodedState, _ := json.Marshal(s)
	node.SetState(encodedState)
}

func getState(node lib.Node) state {
	var s state
	json.Unmarshal(node.GetState(), &s)
	return s
}

func receive(node lib.Node) (message, int) {
	s := getState(node)

	var msg message
	i, receivedMsg := node.ReceiveAnyMessage()
	json.Unmarshal(receivedMsg, &msg)

	bit := msg.Bit
	if i == 0 {
		s.Messages.LeftMessages[bit] = msg
	} else {
		s.Messages.RightMessages[bit] = msg
	}

	setState(node, s)
	return msg, i
}

func send(node lib.Node, direction int, id int, hop int, bit int) {
	m := message{Id: id, Hop: hop, Bit: bit}
	msg, _ := json.Marshal(m)
	node.SendMessage(direction, msg)
}

func initialize(node lib.Node, domainSize int) {
	id := rand.Intn(domainSize)
	emptyMessage := newEmptyMessage()
	messages := storedMessages{
		LeftMessages:  [2]message{emptyMessage, emptyMessage},
		RightMessages: [2]message{emptyMessage, emptyMessage},
	}
	state := state{
		Id:         id,
		Status:     active,
		DomainSize: domainSize,
		Messages:   messages,
		Bit:        0,
	}

	setState(node, state)
}

func startNewRound(node lib.Node) {
	s := getState(node)

	bit := s.Bit
	s.Messages.LeftMessages[bit] = newEmptyMessage()
	s.Messages.RightMessages[bit] = newEmptyMessage()

	bit = 1 - bit
	s.Bit = bit

	id := rand.Intn(s.DomainSize)
	s.Id = id

	setState(node, s)
	send(node, 0, id, 1, bit)
	send(node, 1, id, 1, bit)
}

func runActive(node lib.Node) int {
	s := getState(node)

	send(node, 0, s.Id, 1, s.Bit)
	send(node, 1, s.Id, 1, s.Bit)

	for {
		s := getState(node)

		bit := s.Bit
		if !isMessageEmpty(s.Messages.LeftMessages[bit]) && !isMessageEmpty(s.Messages.RightMessages[bit]) {
			msgLeft := s.Messages.LeftMessages[bit]
			msgRight := s.Messages.RightMessages[bit]

			if msgLeft.Hop == node.GetSize() {
				// node becomes a leader
				s.Status = leader
				setState(node, s)
				return 1
			}

			if msgLeft.Id > s.Id || msgRight.Id > s.Id {
				// node becomes passive
				if !isMessageEmpty(s.Messages.LeftMessages[1-bit]) {
					m := s.Messages.LeftMessages[1-bit]
					send(node, 1, m.Id, m.Hop+1, m.Bit)
					s.Messages.LeftMessages[1-bit] = newEmptyMessage()
				}
				if !isMessageEmpty(s.Messages.RightMessages[1-bit]) {
					m := s.Messages.RightMessages[1-bit]
					send(node, 0, m.Id, m.Hop+1, m.Bit)
					s.Messages.RightMessages[1-bit] = newEmptyMessage()
				}

				s.Status = passive
				setState(node, s)

				return 0
			}

			startNewRound(node)
			continue
		}

		receive(node)
	}
}

func runLeader(node lib.Node) {
	m := newLeaderMessage()
	send(node, 0, m.Id, m.Hop, m.Bit)

	for {
		msg, _ := receive(node)
		if isMessageFromLeader(msg) {
			log.Println("Node", node.GetIndex(), "is a leader")
			return
		}
	}
}

func runPassive(node lib.Node) {
	for {
		msg, direction := receive(node)
		direction = 1 - direction
		msg.Hop += 1
		send(node, direction, msg.Id, msg.Hop, msg.Bit)

		if isMessageFromLeader(msg) {
			var s state
			json.Unmarshal(node.GetState(), &s)
			s.Status = nonleader
			setState(node, s)

			log.Println("Node", node.GetIndex(), "is not a leader")
			return
		}
	}
}

func run(node lib.Node, domainSize int) {
	node.StartProcessing()
	initialize(node, domainSize)

	status := runActive(node)
	if status == 1 {
		runLeader(node)
	} else {
		runPassive(node)
	}
	node.FinishProcessing(true)
}

func check(nodes []lib.Node) {
	leaders := 0
	for _, v := range nodes {
		s := getState(v)

		if s.Status == active || s.Status == passive {
			panic(fmt.Sprint("Node", v.GetIndex(), "has status: ", s.Status))
		} else if s.Status == leader {
			leaders++
		}
	}

	if leaders != 1 {
		panic(fmt.Sprint("There are", leaders, "leaders"))
	}
}

func Run(n int, domainSize int) (int, int) {
	rand.Seed(time.Now().UnixNano())

	nodes, runner := lib.BuildRing(n)
	for _, node := range nodes {
		log.Println("Node", node.GetIndex(), "is starting")
		go run(node, domainSize)
	}

	runner.Run(true)
	check(nodes)

	return runner.GetStats()
}
