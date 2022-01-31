package directed_ring

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

const CLOSING_ROUND = -1

type stateAsyncItaiRodeh struct {
	Id               int
	State            APL
	Round            int
	MessagesSent     int
	MessagesReceived int
	ClosingMode      bool
	ClosingMessageId int
}

type messageAsyncItaiRodeh struct {
	Id        int
	Round     int
	Hop       int
	Bit       bool
	MessageId int
}

func setStateAsyncItaiRodeh(v lib.Node, state stateAsyncItaiRodeh) {
	data, _ := json.Marshal(state)
	v.SetState(data)
}

func getStateAsyncItaiRodeh(node lib.Node) stateAsyncItaiRodeh {
	var state stateAsyncItaiRodeh
	json.Unmarshal(node.GetState(), &state)
	return state
}

func sendAsyncItaiRodeh(v lib.Node, state *stateAsyncItaiRodeh, id int, round int, hop int, bit bool) {
	(*state).MessagesSent++
	setStateAsyncItaiRodeh(v, *state)

	outMessage, _ := json.Marshal(messageAsyncItaiRodeh{
		Id:        id,
		Round:     round,
		Hop:       hop,
		Bit:       bit,
		MessageId: (*state).MessagesSent,
	})
	go v.SendMessage(0, outMessage)
}

func receiveAsyncItaiRodeh(v lib.Node) (messageAsyncItaiRodeh, stateAsyncItaiRodeh) {
	var message messageAsyncItaiRodeh
	inMessage := v.ReceiveMessage(0)
	json.Unmarshal(inMessage, &message)

	state := getStateAsyncItaiRodeh(v)
	state.MessagesReceived++
	setStateAsyncItaiRodeh(v, state)

	return message, state
}

func initializeAsyncItaiRodeh(v lib.Node) {
	setStateAsyncItaiRodeh(v, stateAsyncItaiRodeh{
		Id:               rand.Int(),
		State:            ACTIVE,
		Round:            1,
		MessagesSent:     0,
		MessagesReceived: 0,
		ClosingMode:      false,
		ClosingMessageId: 0,
	})
	state := getStateAsyncItaiRodeh(v)
	sendAsyncItaiRodeh(v, &state, state.Id, state.Round, 1, true)
}

func processAsyncItaiRodeh(v lib.Node) bool {
	message, state := receiveAsyncItaiRodeh(v)

	if message.Round == CLOSING_ROUND {
		state.ClosingMode = true
		state.ClosingMessageId = message.MessageId
		setStateAsyncItaiRodeh(v, state)
		if state.State != LEADER {
			sendAsyncItaiRodeh(v, &state, message.Id, CLOSING_ROUND, 0, false)
		}
	}

	if state.ClosingMode {
		if state.MessagesReceived == state.ClosingMessageId {
			return false
		}
	} else {
		if state.State == PASSIVE {
			sendAsyncItaiRodeh(v, &state, message.Id, message.Round, message.Hop+1, message.Bit)
		} else if state.State == ACTIVE {
			if message.Hop == v.GetSize() && message.Bit {
				state.State = LEADER
				state.ClosingMode = true
				sendAsyncItaiRodeh(v, &state, message.Id, CLOSING_ROUND, 1, false)
			} else if message.Hop == v.GetSize() && !message.Bit {
				state.Id = rand.Int()
				state.Round++
				sendAsyncItaiRodeh(v, &state, state.Id, state.Round, 1, true)
			} else if message.Round == state.Round && message.Id == state.Id {
				sendAsyncItaiRodeh(v, &state, message.Id, message.Round, message.Hop+1, false)
			} else if message.Round > state.Round || message.Round == state.Round && message.Id > state.Id {
				state.State = PASSIVE
				sendAsyncItaiRodeh(v, &state, message.Id, message.Round, message.Hop+1, message.Bit)
			} // else purge message
		}
	}
	return true
}

func runAsyncItaiRodeh(v lib.Node) {
	v.StartProcessing()
	initializeAsyncItaiRodeh(v)
	for processAsyncItaiRodeh(v) {
	}
	v.FinishProcessing(true)
}

func checkSingleLeaderElected(vertices []lib.Node) bool {
	var leaders int = 0
	for _, v := range vertices {
		state := getStateAsyncItaiRodeh(v)
		if state.State == LEADER {
			leaders++
		}
	}
	return leaders == 1
}

func RunAsyncItaiRodeh(n int) (int, int) {
	rand.Seed(time.Now().UnixNano())

	vertices, runner := lib.BuildDirectedRing(n)

	for _, v := range vertices {
		log.Println("Running node", v.GetIndex())
		go runAsyncItaiRodeh(v)
	}

	runner.Run(false)
	log.Println("Single leader selected: ", checkSingleLeaderElected(vertices))
	return runner.GetStats()
}

func wtf() {

}
