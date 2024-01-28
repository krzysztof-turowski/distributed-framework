package sync_metivier_a

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

const (
	None = iota
)

type State struct {
	InGraph        bool
	InMIS          bool
	NeighborInMIS  bool
	SubPhaseNum    int
	RandomMin      uint64
	ReceivedMin    uint64
	AliveNeighbors map[int]bool
}

const (
	phaseMin = iota
	phaseClaim
	phaseDecline
	subPhases = 3
)

type Message struct {
	MessageType int
	Info        uint64
}

const (
	SendingMinimum = iota
	JoiningMIS
	NeverJoiningMIS
)

func setState(node lib.Node, state State) {
	newState, _ := json.Marshal(state)
	node.SetState(newState)
}

func getState(node lib.Node) State {
	encodedState := node.GetState()
	var state State
	json.Unmarshal(encodedState, &state)
	return state
}

func prepareMessage(message Message) []byte {
	encoded, _ := json.Marshal(message)
	return encoded
}

func receiveMessage(node lib.Node, idx int) Message {
	var message Message
	receivedMessage := node.ReceiveMessage(idx)
	err := json.Unmarshal(receivedMessage, &message)
	if err != nil {
		return Message{}
	}
	return message
}

func sendNullMessage(node lib.Node, i int) {
	node.SendMessage(i, nil)
}

func initialize(node lib.Node) {
	state := State{
		true,
		false,
		false,
		None,
		None,
		None,
		map[int]bool{},
	}
	for i := range node.GetInNeighbors() {
		state.AliveNeighbors[i] = true
	}
	setState(node, state)
}

func phasedSend(
	node lib.Node,
	state *State,
	subPhaseNum int,
	idx int,
) {
	switch subPhaseNum {
	case phaseMin:
		node.SendMessage(
			idx,
			prepareMessage(
				Message{
					SendingMinimum,
					state.RandomMin,
				},
			))
	case phaseClaim:
		if state.InMIS {
			node.SendMessage(
				idx,
				prepareMessage(
					Message{
						JoiningMIS,
						state.RandomMin,
					},
				))
		} else {
			sendNullMessage(node, idx)
		}
	case phaseDecline:
		if state.NeighborInMIS {
			node.SendMessage(
				idx,
				prepareMessage(
					Message{
						NeverJoiningMIS,
						state.RandomMin,
					},
				))
			state.InGraph = false
		} else {
			sendNullMessage(node, idx)
		}
	}
}

func phasedReceive(
	node lib.Node,
	nodeState *State,
	subPhaseNum int,
	idx int,
) {
	message := receiveMessage(node, idx)
	switch subPhaseNum {
	case phaseMin:
		nodeState.ReceivedMin = min(nodeState.ReceivedMin, message.Info)
	case phaseClaim:
		if message.MessageType == JoiningMIS {
			nodeState.NeighborInMIS = true
		}
	case phaseDecline:
		if message.MessageType == NeverJoiningMIS {
			nodeState.AliveNeighbors[idx] = false
		}
	}
}

func process(node lib.Node, subPhaseNum int) bool {
	state := getState(node)
	for i := range node.GetInNeighbors() {
		if state.AliveNeighbors[i] {
			phasedSend(node, &state, subPhaseNum, i)
		}

	}
	for i := range node.GetInNeighbors() {
		if state.AliveNeighbors[i] {
			phasedReceive(node, &state, subPhaseNum, i)
		}

	}
	if subPhaseNum == phaseMin {
		state.InMIS = state.RandomMin < state.ReceivedMin
	}
	if subPhaseNum == phaseDecline {
		state.InGraph = !state.InMIS && !state.NeighborInMIS
	}
	setState(node, state)
	return !state.InGraph
}

func uniform() uint64 {
	return rand.Uint64()
}

func check(nodes []lib.Node) {
	for i, node := range nodes {
		state := getState(node)
		if state.InMIS != true && state.InMIS != false {
			panic(fmt.Sprint("Incorrect location at ", i))
		}

		outCnt := 0
		inCnt := 0

		for j, neighbor := range node.GetInNeighbors() {
			if getState(neighbor).InMIS == true {
				inCnt++
			} else {
				log.Println(j, " neighbor of ", i)
				outCnt++
			}
		}

		if state.InMIS == true && inCnt > 0 {
			panic(fmt.Sprint(i, " and his neighbor ended up in MIS"))
		}
		if state.InMIS == false && inCnt == 0 {
			panic(fmt.Sprint(i, " and his neighbor did not end up in MIS"))
		}
	}
}

func Run(n int, p float64) (int, int) {
	nodes, synchronizer := lib.BuildSynchronizedRandomGraph(n, p)
	for _, node := range nodes {
		go run(node)
	}
	synchronizer.Synchronize(0)
	check(nodes)
	return synchronizer.GetStats()
}

func run(node lib.Node) {
	node.StartProcessing()
	initialize(node)
	finish := false
	node.FinishProcessing(finish)
	for !finish {
		node.StartProcessing()
		state := getState(node)
		state.RandomMin = uniform()
		state.ReceivedMin = math.MaxUint64
		setState(node, state)
		for i := 0; i < subPhases; i++ {
			finish = finish || process(node, i)
		}
		node.FinishProcessing(finish)
	}
}
