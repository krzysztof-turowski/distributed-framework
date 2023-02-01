package async_loui_matsushita_west

import (
	"encoding/json"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
)

type MessageType int

const (
	msgElect MessageType = iota
	msgLeader
)

const none = -1

type message struct {
	Type         MessageType
	Id           int
	DirectedDist int
}

type state struct {
	Active         bool
	LastActiveDist int
	NumberOfNodes  int
	Leader         int
}

func getState(node lib.Node) state {
	var state state
	json.Unmarshal(node.GetState(), &state)
	return state
}

func setState(node lib.Node, state *state) {
	encodedState, _ := json.Marshal(*state)
	node.SetState(encodedState)
}

func send(
	node lib.Node,
	index int,
	id int,
	distance int,
) {
	outMessage, _ := json.Marshal(message{
		Type:         msgElect,
		Id:           id,
		DirectedDist: distance,
	})
	node.SendMessage(index, outMessage)
}

func broadcastLeader(node lib.Node) {
	outMessage, _ := json.Marshal(message{
		Type:         msgLeader,
		Id:           node.GetIndex(),
		DirectedDist: 0,
	})
	for i := 0; i < node.GetOutChannelsCount(); i++ {
		node.SendMessage(i, outMessage)
	}
}

func receive(node lib.Node) message {
	var msg message
	_, inMessage := node.ReceiveAnyMessage()
	json.Unmarshal(inMessage, &msg)
	return msg
}

func initialize(node lib.Node) {
	n := node.GetOutChannelsCount() + 1
	setState(node, &state{
		Active:         true,
		LastActiveDist: n - 1,
		NumberOfNodes:  n,
		Leader:         none,
	})

	id := node.GetIndex()
	send(node, 0, id, 1)
}

func process(node lib.Node) bool {
	state := getState(node)
	msg := receive(node)
	id := node.GetIndex()
	defer setState(node, &state)

	switch msg.Type {
	case msgElect:
		if state.Active {
			if msg.Id == id {
				state.Leader = id
				broadcastLeader(node)
				return false
			} else if msg.Id < id {
				distance := state.NumberOfNodes - msg.DirectedDist
				send(node, distance-1, id, distance)
			} else {
				state.Active = false
				state.LastActiveDist = msg.DirectedDist
			}
		} else {
			nextNodeIndex := state.NumberOfNodes - state.LastActiveDist - 1
			distance := msg.DirectedDist - state.LastActiveDist
			if distance < 0 {
				distance += state.NumberOfNodes
			}
			send(node, nextNodeIndex, msg.Id, distance)
		}
	case msgLeader:
		state.Leader = msg.Id
		return false
	}

	return true
}

func run(node lib.Node) {
	node.StartProcessing()
	initialize(node)

	for process(node) {
	}

	node.FinishProcessing(true)
}

func Run(nodes []lib.Node, runner lib.Runner) (int, int) {
	for _, node := range nodes {
		log.Println("Running node", node.GetIndex())
		go run(node)
	}

	runner.Run(true)
	checkSingleLeaderElected(nodes)

	return runner.GetStats()
}

func checkSingleLeaderElected(nodes []lib.Node) {
	leader := getState(nodes[0]).Leader
	if leader == none {
		panic("No leader elected")
	}
	for _, node := range nodes {
		if getState(node).Leader != leader {
			panic("Multiple leaders elected")
		}
	}
	log.Println("Elected node", leader, "as a leader")
}
