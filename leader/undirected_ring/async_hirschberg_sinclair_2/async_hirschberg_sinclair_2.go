package async_hirschberg_sinclair_2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

const (
	out       = iota
	candidate = iota
	leader    = iota
	over      = iota
)

const (
	from        = iota
	no          = iota
	ok          = iota
	leaderfound = iota
)

type message struct {
	content   []byte
	direction int
}

type state struct {
	id          int
	phase       int
	Status      int
	confirm     int
	outMessages chan message
}

type messageContent struct {
	MessegeType int
	Content     []int
}

func send(node lib.Node, m messageContent, direction int, waitGroup *sync.WaitGroup, outMessages chan message) {
	bytesMessage, _ := json.Marshal(m)
	waitGroup.Add(1)
	outMessages <- message{
		content:   bytesMessage,
		direction: direction,
	}
}

func recive(node lib.Node) (messageContent, int) {
	direction, bytesMessage := node.ReceiveAnyMessage()
	var message messageContent
	json.Unmarshal(bytesMessage, &message)
	return message, direction
}

func sendBack(node lib.Node, m messageContent, direction int, waitGroup *sync.WaitGroup, outMessages chan message) {
	send(node, m, direction, waitGroup, outMessages)
}

func sendForward(node lib.Node, m messageContent, direction int, waitGroup *sync.WaitGroup, outMessages chan message) {
	send(node, m, (direction ^ 1), waitGroup, outMessages)
}

func startNewStage(node lib.Node, state *state, waitGroup *sync.WaitGroup) {
	state.phase++
	state.confirm = 0
	messageContent := messageContent{
		MessegeType: from,
		Content:     []int{state.id, 0, (1 << state.phase)},
	}
	send(node, messageContent, 0, waitGroup, state.outMessages)
	send(node, messageContent, 1, waitGroup, state.outMessages)
}

func handleMessage(node lib.Node, content messageContent, direction int, state *state, waitGroup *sync.WaitGroup) {
	messageStatus := content.MessegeType
	if messageStatus == from {
		senderId := content.Content[0]
		stepCounter := content.Content[1] + 1
		StepLimit := content.Content[2]
		if senderId == state.id {
			if state.Status != leader {
				state.Status = leader
				send(node, messageContent{
					MessegeType: leaderfound,
					Content:     []int{stepCounter},
				}, 0, waitGroup, state.outMessages)
			}
			return
		} else if senderId < state.id {
			sendBack(node, messageContent{
				MessegeType: no,
				Content:     []int{senderId},
			}, direction, waitGroup, state.outMessages)
		} else if senderId > state.id {
			state.Status = out
			if stepCounter == StepLimit {
				sendBack(node, messageContent{
					MessegeType: ok,
					Content:     []int{senderId},
				}, direction, waitGroup, state.outMessages)
			} else {
				sendForward(node, messageContent{
					MessegeType: from,
					Content:     []int{senderId, stepCounter, StepLimit},
				}, direction, waitGroup, state.outMessages)
			}
		}
		return
	}
	if messageStatus == ok || messageStatus == no {
		adressant := content.Content[0]
		if adressant != state.id {
			sendForward(node, content, direction, waitGroup, state.outMessages)
		} else {
			if messageStatus == no {
				state.Status = out
			}
			state.confirm++
			if state.confirm == 2 && state.Status == candidate {
				startNewStage(node, state, waitGroup)
			}
		}
		return
	}
	if messageStatus == leaderfound && content.Content[0] != 1 {
		state.Status = over
		sendForward(node, messageContent{
			MessegeType: leaderfound,
			Content:     []int{content.Content[0] - 1},
		}, direction, waitGroup, state.outMessages)
		return
	}
}

func processRun(node lib.Node, WaitGroup *sync.WaitGroup) {
	state := state{
		id:          node.GetIndex(),
		phase:       -1,
		Status:      candidate,
		confirm:     0,
		outMessages: make(chan message, 10000),
	}

	go func() {
		for msg := range state.outMessages {
			node.SendMessage(msg.direction, msg.content)
		}
	}()

	node.StartProcessing()
	startNewStage(node, &state, WaitGroup)
	for {
		messageContent, direction := recive(node)
		handleMessage(node, messageContent, direction, &state, WaitGroup)
		if state.Status == leader || state.Status == over {
			WaitGroup.Done()
			break
		}
	}

	go func() {
		for {
			recive(node)
		}
	}()

	WaitGroup.Wait()

	bytesState, _ := json.Marshal(state)
	node.SetState(bytesState)
	node.FinishProcessing(true)
	if state.Status == leader {
		log.Println("Node", node.GetIndex(), "is a selected leader")
	}
}

func resultChecker(vertices []lib.Node) {
	leaderCount, leaderMax, realMax := 0, 0, 0
	for _, vertex := range vertices {
		state := state{}
		json.Unmarshal(vertex.GetState(), &state)
		if state.Status == leader {
			leaderCount++
			leaderMax = vertex.GetIndex()
		}
		realMax = max(realMax, vertex.GetIndex())
	}
	if leaderCount == 0 {
		panic("No leader elected")
	}
	if leaderCount != 1 {
		panic("Multiple leaders elected")

	}
	if leaderMax != realMax {
		panic("Leader is not the maximum")
	}
}

func Run(n int) (int, int) {
	log.SetOutput(ioutil.Discard)
	for i := 0; i < 100; i++ {
		fmt.Println("Iteration", i)
		WaitGroup := sync.WaitGroup{}
		WaitGroup.Add(n)
		vertices, runner := lib.BuildRing(n)
		for _, vertex := range vertices {
			go processRun(vertex, &WaitGroup)
		}
		runner.Run(true)
		WaitGroup.Wait()
		resultChecker(vertices)
	}
	return 0, 0
}
