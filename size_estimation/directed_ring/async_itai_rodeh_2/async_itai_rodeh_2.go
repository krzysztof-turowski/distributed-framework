package async_itai_rodeh_2

import (
	"encoding/json"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
	"math/rand"
	"time"
)

type State struct {
	Size               int
	Id                 bool
	Confidence         int
	RequiredConfidence int
}

type Message struct {
	K     int
	Id    bool
	Count int
}

func setState(node lib.Node, state *State) {
	encoded, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	node.SetState(encoded)
}

func getState(node lib.Node) State {
	var state State
	err := json.Unmarshal(node.GetState(), &state)
	if err != nil {
		panic(err)
	}
	return state
}

func receive(node lib.Node) (Message, State) {
	inMessage := node.ReceiveMessageWithTimeout(0)
	var msg Message
	if inMessage != nil {
		err := json.Unmarshal(inMessage, &msg)
		if err != nil {
			panic(err)
		}
	}
	return msg, getState(node)
}

func send(node lib.Node, state State, message Message) {
	setState(node, &state)
	outMessage, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}
	go node.SendMessage(0, outMessage)
}

func initialize(node lib.Node) {
	state := getState(node)
	send(node, state, Message{
		K:     2,
		Id:    getBit(),
		Count: 0,
	})
}

func getBit() bool {
	return rand.Intn(2) != 0
}

func process(node lib.Node) bool {
	message, state := receive(node)
	if message.K == 0 {
		return false
	}
	message.Count += 1
	if message.K > message.Count {
		if message.K > state.Size {
			state.Size = message.K
			state.Confidence = 0
			state.Id = getBit()
			msgOut := Message{
				K:     state.Size,
				Id:    state.Id,
				Count: 0,
			}
			send(node, state, msgOut)
		}
		send(node, state, message)
		return true
	}
	if message.K < state.Size {
		return true
	}
	if message.K > state.Size || message.Id != state.Id || state.Confidence == state.RequiredConfidence {
		state.Size = message.K + 1
		state.Confidence = 0
		state.Id = getBit()
		msgOut := Message{
			K:     state.Size,
			Id:    state.Id,
			Count: 0,
		}
		send(node, state, msgOut)
		return true
	}
	state.Confidence += 1
	state.Id = getBit()
	if state.Confidence < state.RequiredConfidence {
		msgOut := Message{
			K:     state.Size,
			Id:    state.Id,
			Count: 0,
		}
		send(node, state, msgOut)
	} else {
		log.Println("Node", node.GetIndex(), "finished with size", state.Size)
	}
	return true
}

func Run(numOfProcessors, confidence int) (int, int) {
	rand.Seed(time.Now().UnixNano())
	nodes, runner := lib.BuildDirectedRing(numOfProcessors)
	for _, node := range nodes {
		state := State{
			Id:                 false,
			Size:               2,
			Confidence:         0,
			RequiredConfidence: confidence,
		}
		setState(node, &state)
	}
	for _, node := range nodes {
		log.Println("Running node", node.GetIndex())
		go run(node)
	}

	runner.Run(false)
	allSame, correct, result := checkResult(nodes)
	log.Println("All results are the same:", allSame)
	log.Println("Result is correct:", correct)
	log.Println("Result:", result)
	return runner.GetStats()
}

func run(node lib.Node) {
	node.StartProcessing()
	initialize(node)
	for process(node) {
	}
	node.FinishProcessing(true)
}

func checkResult(nodes []lib.Node) (bool, bool, int) {
	var allSame, correct = true, true
	var result = getState(nodes[0]).Size
	for _, node := range nodes {
		nodeResult := getState(node).Size
		if nodeResult != result {
			allSame = false
			break
		}
	}
	if result != len(nodes) {
		correct = false
	}
	return allSame, correct, result
}
