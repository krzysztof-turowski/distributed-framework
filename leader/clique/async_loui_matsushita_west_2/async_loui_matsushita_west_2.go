package async_loui_matsushita_west_2

import (
	"encoding/json"
	"fmt"
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

const (
	None = iota
)

type State struct {
	Status int
	Leader int
	N      int
	D      int
	E      int
}

const (
	Active = iota
	Passive
)

type Message struct {
	Sender      int
	MessageType int
	Distance    int
}

const (
	DeclaringLeader = iota
	Electing
)

func setState(node lib.Node, state *State) {
	encodedState, _ := json.Marshal(*state)
	node.SetState(encodedState)
}

func getState(node lib.Node) State {
	var state State
	err := json.Unmarshal(node.GetState(), &state)
	if err != nil {
		panic(fmt.Sprint("Error while decoding data on node ", node.GetIndex()))
		return State{}
	}
	return state
}

func prepareMessage(message Message) []byte {
	encoded, _ := json.Marshal(message)
	return encoded
}

func receiveMessage(node lib.Node) Message {
	var message Message
	_, receivedMessage := node.ReceiveAnyMessage()
	err := json.Unmarshal(receivedMessage, &message)
	if err != nil {
		panic(fmt.Sprint("Error while decoding data on node ", node.GetIndex()))
		return Message{}
	}
	return message
}

func sendNullMessage(node lib.Node, i int) {
	node.SendMessage(i, nil)
}

func initialize(node lib.Node) {
	n := node.GetSize() + 1
	setState(
		node,
		&State{
			Active,
			None,
			n,
			None,
			None,
		},
	)

	id := node.GetIndex()
	fmt.Println("Node: ", id)
	node.SendMessage(0, prepareMessage(Message{id, Electing, 1}))
}

func process(node lib.Node) bool {
	id := node.GetIndex()
	message := receiveMessage(node)
	state := getState(node)
	defer setState(node, &state)

	switch message.MessageType {
	case DeclaringLeader:
		state.Leader = message.Sender
	case Electing:
		switch state.Status {
		case Active:
			if message.Sender == id {
				state.Leader = id
				encoded := prepareMessage(
					Message{
						id,
						DeclaringLeader,
						state.N - message.Distance,
					},
				)
				for i := 0; i < node.GetOutChannelsCount(); i++ {
					node.SendMessage(
						i,
						encoded,
					)
				}
			} else if message.Sender > id {
				state.Status = Passive
				state.E = message.Distance
			} else {
				node.SendMessage(
					state.N-message.Distance-1,
					prepareMessage(
						Message{
							id,
							Electing,
							state.N - message.Distance,
						},
					),
				)
			}
		case Passive:
			node.SendMessage(
				state.N-state.E-1,
				prepareMessage(
					Message{
						message.Sender,
						Electing,
						message.Distance - state.E,
					},
				),
			)
		}
	}

	return state.Leader != None
}

func checkSingleLeaderElected(nodes []lib.Node) {
	leader := getState(nodes[0]).Leader
	if leader == None {
		panic(fmt.Sprint("Error due to lack of leader"))
	}
	for _, node := range nodes {
		if getState(node).Leader != leader {
			panic(fmt.Sprint("Error due to leader ambiguity"))
		}
	}
	fmt.Println("Leader: ", leader)
}

func Run(nodes []lib.Node, runner lib.Runner) (int, int) {
	for _, node := range nodes {
		go run(node)
	}
	runner.Run(true)
	checkSingleLeaderElected(nodes)
	return runner.GetStats()
}

func run(node lib.Node) {
	node.StartProcessing()
	initialize(node)
	for process(node) {
	}
	node.FinishProcessing(true)
}
