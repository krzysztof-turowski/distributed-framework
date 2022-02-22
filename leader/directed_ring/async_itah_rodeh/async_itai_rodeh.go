package async_itah_rodeh

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

type APL byte

const (
	ACTIVE APL = iota
	PASSIVE
	LEADER
)

const ClosingRound = -1

type state struct {
	Id               int
	State            APL
	Round            int
	MessagesSent     int
	MessagesReceived int
	ClosingMode      bool
	ClosingMessageId int
}

type message struct {
	Id        int
	Round     int
	Hop       int
	Bit       bool
	MessageId int
}

func setState(v lib.Node, state state) {
	data, _ := json.Marshal(state)
	v.SetState(data)
}

func getState(node lib.Node) state {
	var state state
	json.Unmarshal(node.GetState(), &state)
	return state
}

func send(v lib.Node, state *state, id int, round int, hop int, bit bool) {
	(*state).MessagesSent++
	setState(v, *state)

	outMessage, _ := json.Marshal(message{
		Id:        id,
		Round:     round,
		Hop:       hop,
		Bit:       bit,
		MessageId: (*state).MessagesSent,
	})
	go v.SendMessage(0, outMessage)
}

func receive(v lib.Node) (message, state) {
	var message message
	inMessage := v.ReceiveMessage(0)
	json.Unmarshal(inMessage, &message)

	state := getState(v)
	state.MessagesReceived++
	setState(v, state)

	return message, state
}

func initialize(v lib.Node) {
	setState(v, state{
		Id:               rand.Int(),
		State:            ACTIVE,
		Round:            1,
		MessagesSent:     0,
		MessagesReceived: 0,
		ClosingMode:      false,
		ClosingMessageId: 0,
	})
	state := getState(v)
	send(v, &state, state.Id, state.Round, 1, true)
}

func process(v lib.Node) bool {
	message, state := receive(v)

	if message.Round == ClosingRound {
		state.ClosingMode = true
		state.ClosingMessageId = message.MessageId
		setState(v, state)
		if state.State != LEADER {
			send(v, &state, message.Id, ClosingRound, 0, false)
		}
	}

	if state.ClosingMode {
		if state.MessagesReceived == state.ClosingMessageId {
			return false
		}
	} else {
		if state.State == PASSIVE {
			send(v, &state, message.Id, message.Round, message.Hop+1, message.Bit)
		} else if state.State == ACTIVE {
			if message.Hop == v.GetSize() && message.Bit {
				state.State = LEADER
				state.ClosingMode = true
				send(v, &state, message.Id, ClosingRound, 1, false)
			} else if message.Hop == v.GetSize() && !message.Bit {
				state.Id = rand.Int()
				state.Round++
				send(v, &state, state.Id, state.Round, 1, true)
			} else if message.Round == state.Round && message.Id == state.Id {
				send(v, &state, message.Id, message.Round, message.Hop+1, false)
			} else if message.Round > state.Round || message.Round == state.Round && message.Id > state.Id {
				state.State = PASSIVE
				send(v, &state, message.Id, message.Round, message.Hop+1, message.Bit)
			} // else purge message
		}
	}
	return true
}

func run(v lib.Node) {
	v.StartProcessing()
	initialize(v)
	for process(v) {
	}
	v.FinishProcessing(true)
}

func checkSingleLeaderElected(vertices []lib.Node) bool {
	var leaders int = 0
	for _, v := range vertices {
		state := getState(v)
		if state.State == LEADER {
			leaders++
		}
	}
	return leaders == 1
}

func Run(n int) (int, int) {
	rand.Seed(time.Now().UnixNano())

	vertices, runner := lib.BuildDirectedRing(n)

	for _, v := range vertices {
		log.Println("Running node", v.GetIndex())
		go run(v)
	}

	runner.Run(false)
	log.Println("Single leader selected: ", checkSingleLeaderElected(vertices))
	return runner.GetStats()
}
