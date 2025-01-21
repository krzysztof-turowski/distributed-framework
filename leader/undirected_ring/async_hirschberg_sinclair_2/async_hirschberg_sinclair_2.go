package async_hirschberg_sinclair_2

import (
	"encoding/json"
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
	finished    = iota
)

type state struct {
	id            int
	phase         int
	Status        int
	confirm       int
	waitingToSent int
	mu            sync.Mutex
}

type messageContent struct {
	MessegeType int
	Content     []int
}

func send(m messageContent, direction int, state *state, node lib.Node) {
	bytesMessage, _ := json.Marshal(m)
	state.mu.Lock()
	state.waitingToSent++
	state.mu.Unlock()
	go func() {
		defer func() {
			state.mu.Lock()
			state.waitingToSent--
			state.mu.Unlock()
		}()
		node.SendMessage(direction, bytesMessage)
	}()

}

func recive(node lib.Node) (messageContent, int) {
	direction, bytesMessage := node.ReceiveAnyMessage()
	var message messageContent
	json.Unmarshal(bytesMessage, &message)
	return message, direction
}

func sendBack(m messageContent, direction int, state *state, node lib.Node) {
	send(m, direction, state, node)
}

func sendForward(m messageContent, direction int, state *state, node lib.Node) {
	send(m, (direction ^ 1), state, node)
}

func startNewStage(state *state, node lib.Node) {
	state.phase++
	state.confirm = 0
	messageContent := messageContent{
		MessegeType: from,
		Content:     []int{state.id, 0, (1 << state.phase)},
	}
	send(messageContent, 0, state, node)
	send(messageContent, 1, state, node)
}

func handleMessage(content messageContent, direction int, state *state, node lib.Node) {
	messageStatus := content.MessegeType
	if state.Status != over && state.Status != leader {
		if messageStatus == from {
			senderId := content.Content[0]
			stepCounter := content.Content[1] + 1
			StepLimit := content.Content[2]
			if senderId == state.id {
				if state.Status != leader {
					state.confirm = 0
					state.Status = leader
					send(messageContent{
						MessegeType: leaderfound,
						Content:     []int{senderId},
					}, 0, state, node)
				}
				return
			} else if senderId < state.id {
				sendBack(messageContent{
					MessegeType: no,
					Content:     []int{senderId},
				}, direction, state, node)
			} else if senderId > state.id {
				state.Status = out
				if stepCounter == StepLimit {
					sendBack(messageContent{
						MessegeType: ok,
						Content:     []int{senderId},
					}, direction, state, node)
				} else {
					sendForward(messageContent{
						MessegeType: from,
						Content:     []int{senderId, stepCounter, StepLimit},
					}, direction, state, node)
				}
			}
			return
		}
		if messageStatus == ok || messageStatus == no {
			adressant := content.Content[0]
			if adressant != state.id {
				sendForward(content, direction, state, node)
			} else {
				if messageStatus == no {
					state.Status = out
				}
				state.confirm++
				if state.confirm == 2 && state.Status == candidate {
					startNewStage(state, node)
				}
			}
			return
		}
	} else if messageStatus == finished {
		state.confirm++
		return
	}
	if messageStatus == leaderfound {
		sendBack(messageContent{
			MessegeType: finished,
			Content:     []int{},
		}, direction, state, node)

		if state.Status == leader {
			state.confirm++
		} else {
			state.Status = over
			state.confirm = 1
			sendForward(messageContent{
				MessegeType: leaderfound,
				Content:     []int{state.id},
			}, direction, state, node)
		}

		return
	}

}

func run(node lib.Node) {
	state := state{
		id:            node.GetIndex(),
		phase:         -1,
		Status:        candidate,
		confirm:       0,
		waitingToSent: 0,
	}

	node.StartProcessing()
	startNewStage(&state, node)
	for {
		messageContent, direction := recive(node)
		handleMessage(messageContent, direction, &state, node)
		if (state.Status == leader || state.Status == over) && state.confirm == 2 {
			break
		}
		log.Println("Node", node.GetIndex(), "status", state.Status, "confirm", state.confirm)
	}

	for state.waitingToSent > 0 {
	}

	bytesState, _ := json.Marshal(state)
	node.SetState(bytesState)

	node.FinishProcessing(true)
	if state.Status == leader {
		log.Println("Node", node.GetIndex(), "is a selected leader")
	} else {
		log.Println("Node", node.GetIndex(), "is not a leader")
	}
}

func check(vertices []lib.Node) {
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
	vertices, runner := lib.BuildRing(n)
	for _, vertex := range vertices {
		go run(vertex)
	}
	runner.Run(true)
	check(vertices)
	return runner.GetStats()
}
