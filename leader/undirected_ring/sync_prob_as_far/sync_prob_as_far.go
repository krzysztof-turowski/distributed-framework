package sync_prob_as_far

import (
	"encoding/json"
	"fmt"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
)

type nodeStatusType string
type messageType string

const (
	unknown  nodeStatusType = "unknown"
	follower                = "follower"
	leader                  = "leader"
)

const (
	election messageType = "election"
	notify               = "notify"
)

type state struct {
	Status       nodeStatusType
	Max          int
	OutDirection int
}

type message struct {
	MessageType messageType
	Value       int
}
type rawMessage struct {
	origin int
	data   []byte
}

func (m *rawMessage) isEmpty() bool {
	return len(m.data) == 0
}

func setState(v lib.Node, s state) {
	data, _ := json.Marshal(s)
	v.SetState(data)
}

func getState(v lib.Node) state {
	var s state
	json.Unmarshal(v.GetState(), &s)
	return s
}

func sendMessage(v lib.Node, direction int, msg message) {
	messageData, _ := json.Marshal(msg)
	v.SendMessage(direction, messageData)
}

func initialize(v lib.Node) bool {
	randomGenerator := lib.GetRandomGenerator()
	s := state{Status: unknown, Max: v.GetIndex(), OutDirection: randomGenerator.Int() % 2}
	setState(v, s)
	sendMessage(v, s.OutDirection, message{MessageType: election, Value: v.GetIndex()})
	return false
}

func receiveMessageNonBlocking(v lib.Node, source int) rawMessage {
	return rawMessage{origin: source, data: v.ReceiveMessageIfAvailable(source)}
}

func decodeRawMessage(rawMsg rawMessage) message {
	var msg message
	json.Unmarshal(rawMsg.data, &msg)
	return msg
}
func notifyLeaderElected(v lib.Node) {
	vState := getState(v)
	updateStateStatus(v, leader)
	sendMessage(v, vState.OutDirection, message{MessageType: notify, Value: v.GetIndex()})
}

func routeIfNotify(v lib.Node, rawMsg rawMessage) bool {
	if rawMsg.isEmpty() {
		return false
	}
	msg := decodeRawMessage(rawMsg)
	if msg.MessageType == notify {
		updateStateStatus(v, follower)
		sendMessage(v, rawMsg.origin^1, msg)
		return true
	}
	return false
}

func electOrRouteIfPossible(v lib.Node, msg message, messageOrigin int) {
	vState := getState(v)
	if msg.Value == vState.Max {
		notifyLeaderElected(v)
	} else if msg.Value > vState.Max {
		vState.Max = msg.Value
		setState(v, vState)
		sendMessage(v, messageOrigin^1, msg)
	}
}

func handleMessages(v lib.Node, rawMessage1, rawMessage2 rawMessage) bool {
	if routeIfNotify(v, rawMessage1) || routeIfNotify(v, rawMessage2) {
		return true
	}
	if !rawMessage1.isEmpty() && !rawMessage2.isEmpty() {
		message1, message2 := decodeRawMessage(rawMessage1), decodeRawMessage(rawMessage2)
		if message1.Value > message2.Value {
			electOrRouteIfPossible(v, message1, rawMessage1.origin)
		} else {
			electOrRouteIfPossible(v, message2, rawMessage2.origin)
		}
	} else if !rawMessage1.isEmpty() {
		message1 := decodeRawMessage(rawMessage1)
		electOrRouteIfPossible(v, message1, rawMessage1.origin)
	} else if !rawMessage2.isEmpty() {
		message2 := decodeRawMessage(rawMessage2)
		electOrRouteIfPossible(v, message2, rawMessage2.origin)
	}
	return false
}

func leaderRoutine(v lib.Node) bool {
	vState := getState(v)
	rawMsg := receiveMessageNonBlocking(v, vState.OutDirection^1)
	return decodeRawMessage(rawMsg).MessageType == notify
}

func roundRoutine(v lib.Node) bool {
	if isLeader(v) {
		return leaderRoutine(v)
	}
	rawMessage1 := receiveMessageNonBlocking(v, 0)
	rawMessage2 := receiveMessageNonBlocking(v, 1)
	return handleMessages(v, rawMessage1, rawMessage2)
}

func run(v lib.Node) {
	v.StartProcessing()
	finish := initialize(v)
	v.FinishProcessing(finish)
	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = roundRoutine(v)
		v.FinishProcessing(finish)
	}
}

func Run(n int) (int, int) {
	vertices, synchronizer := lib.BuildSynchronizedRing(n)
	for _, v := range vertices {
		go run(v)
	}
	synchronizer.Synchronize(0)
	verifyLeaderElected(vertices)
	return synchronizer.GetStats()
}

func isLeader(v lib.Node) bool {
	vState := getState(v)
	return vState.Status == leader
}
func verifyLeaderElected(vertices []lib.Node) {
	var leaderNode lib.Node
	var s state
	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if s.Status == leader {
			leaderNode = v
			break
		}
	}
	if leaderNode == nil {
		panic("There is no leader on the undirected undirected_ring")
	}
	max := 0
	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if v.GetIndex() > max {
			max = v.GetIndex()
		}
		if v != leaderNode {
			if s.Status == leader {
				panic(fmt.Sprint(
					"Multiple leaders on the undirected undirected_ring: ", v.GetIndex(), leaderNode.GetIndex()))
			}
			if s.Status != follower {
				panic(fmt.Sprint("Node ", v.GetIndex(), " has state ", s.Status))
			}
		}
	}
	if max != leaderNode.GetIndex() {
		panic(fmt.Sprint("Leader has value ", leaderNode.GetIndex(), " but max is ", max))
	}
	log.Println("LEADER ", leaderNode.GetIndex())
}

func updateStateStatus(v lib.Node, status nodeStatusType) {
	vState := getState(v)
	vState.Status = status
	setState(v, vState)
}
