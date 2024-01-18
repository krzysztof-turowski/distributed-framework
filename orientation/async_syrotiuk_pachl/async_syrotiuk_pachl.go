package async_syrotiuk_pachl

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

const (
	LEFT  = 0
	RIGHT = 1
)

type messageType int

const (
	null messageType = iota
	normal
	end
)

type message struct {
	MsgType  messageType
	Vote     int
	Distance int
}

type state struct {
	Indicative bool
	Agreement  bool
}

func send(v lib.Node, direction int, msgType messageType, vote int, distance int) {
	outMessage, _ := json.Marshal(message{MsgType: msgType, Vote: vote, Distance: distance})
	v.SendMessage(direction, outMessage)
}

func receive(v lib.Node) (int, message) {
	from, inMessage := v.ReceiveAnyMessage()
	var msg message
	json.Unmarshal(inMessage, &msg)
	return from, msg
}

func initialize(v lib.Node) {
	setState(v, &state{
		Indicative: false,
		Agreement:  false,
	})
	go send(v, RIGHT, normal, 1, 1)
}

func process(v lib.Node) bool {
	from, msg := receive(v)
	if msg.MsgType == normal {
		if msg.Distance == v.GetSize() {
			agreement := false
			if msg.Vote > 0 {
				agreement = true
			}
			setState(v, &state{
				Indicative: true,
				Agreement:  agreement,
			})
			go send(v, LEFT, end, 0, 0)
			go send(v, RIGHT, end, 0, 0)
			return false
		}

		if from == LEFT {
			go send(v, RIGHT, normal, msg.Vote+1, msg.Distance+1)
		} else if from == RIGHT && msg.Vote > 0 {
			go send(v, LEFT, normal, msg.Vote-1, msg.Distance+1)
		}
	} else if msg.MsgType == end {
		if from == LEFT {
			go send(v, RIGHT, end, 0, 0)
		} else if from == RIGHT {
			go send(v, LEFT, end, 0, 0)
		}
		return false
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

func check(vertices []lib.Node) {
	var agreementStatus bool
	indicativeCount := 0
	for _, v := range vertices {
		if getState(v).Indicative {
			indicativeCount += 1
			agreementStatus = getState(v).Agreement
		}
	}
	if indicativeCount == 0 {
		panic(fmt.Sprint("There is no indicative node"))
	}
	for _, v := range vertices {
		state := getState(v)
		if state.Indicative && state.Agreement != agreementStatus {
			panic(fmt.Sprint("Indicative nodes have different agreement status"))
		}
	}

	orientation := LEFT
	orientationCount := [2]int{1, 0}
	foundOrientation := -1
	firstNodeIndex := vertices[0].GetIndex()
	previousNodeIndex := vertices[0].GetIndex()
	currentNode := vertices[0].GetInNeighbors()[orientation]
	state := getState(vertices[0])
	if state.Indicative && state.Agreement {
		foundOrientation = orientation
	}
	for currentNode.GetIndex() != firstNodeIndex {
		if currentNode.GetInNeighbors()[orientation].GetIndex() == previousNodeIndex {
			orientation = 1 - orientation
		}
		orientationCount[orientation]++

		state := getState(currentNode)
		if state.Indicative && state.Agreement {
			if foundOrientation == -1 {
				foundOrientation = orientation
			} else if foundOrientation != orientation {
				panic(fmt.Sprint("Indicative nodes have different orientation"))
			}
		}

		previousNodeIndex = currentNode.GetIndex()
		currentNode = currentNode.GetInNeighbors()[orientation]
	}

	if foundOrientation == -1 {
		if orientationCount[LEFT] != orientationCount[RIGHT] {
			panic(fmt.Sprint("There is majority but it was not found"))
		}
	} else {
		if orientationCount[foundOrientation] <= orientationCount[1-foundOrientation] {
			panic(fmt.Sprint("Found orientation is not shared by majority of nodes"))
		}
	}
}

func Run(n int) (int, int) {
	rand.Seed(time.Now().UnixNano())

	vertices, runner := lib.BuildRing(n)

	for _, v := range vertices {
		log.Println("Running node", v.GetIndex())
		go run(v)
	}
	runner.Run(false)
	check(vertices)

	return runner.GetStats()
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
