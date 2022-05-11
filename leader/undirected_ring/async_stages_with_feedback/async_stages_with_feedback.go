package async_stages_with_feedback

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
)

const (
	msgE = iota
	msgA
	msgT
)
const (
	statusECandidate = iota
	statusACandidate
	statusPassive
	statusElected
)
const (
	left = iota
	right
	both
)
const none = -1
const byteNone = 255

/*
	in our model nodes do not run continuously,
	as such status "Available" specified in original algorithm does not exist,
	all nodes initiate the election and so all enter as E-Candidate
*/

type Message struct {
	Type byte
	Val  int64
}

type state struct {
	Status             int
	Messages           []Message
	Leader             int64
	ReverseOrientation bool
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

func initialize(node lib.Node) {
	id := int64(node.GetIndex())
	setState(node, &state{
		Status:   statusECandidate,
		Messages: []Message{{byteNone, none}, {byteNone, none}},
		Leader:   none,
	})

	node.SendMessage(left, encodeAll(byte(msgE), id))
	node.SendMessage(right, encodeAll(byte(msgE), id))
}
func max(a int64, b int64, c int64) int64 {
	s := a
	if b > s {
		s = b
	}
	if c > s {
		s = c
	}
	return s
}
func ops(a int) int {
	if a == right {
		return left
	} else {
		return right
	}
}
func awaitMessage(state state) int {
	if isEmptyMessage(state.Messages[0]) {
		if isEmptyMessage(state.Messages[1]) {
			return both
		}
		return left
	}
	return right //we assume that at least one channel is to be read from
}
func isEmptyMessage(message Message) bool {
	if message.Type == byteNone && message.Val == none {
		return true
	}
	return false
}
func resetMessages(state *state) {
	state.Messages = []Message{{byteNone, none}, {byteNone, none}}
}
func heldMessages(messages []Message) int {
	val := 0
	for _, msg := range messages {
		if msg.Type != byteNone && msg.Val != none {
			val += 1
		}
	}
	return val
}
func process(node lib.Node, sender int, message []byte) bool {
	state := getState(node)
	id := int64(node.GetIndex())
	defer setState(node, &state)
	var msg Message
	decodeAll(message, &msg)
	switchState := state.Status
	switch switchState {
	case statusECandidate:
		if isEmptyMessage(state.Messages[sender]) {
			state.Messages[sender] = msg
		}
		if heldMessages(state.Messages) == 2 {
			vl := state.Messages[left].Val
			vr := state.Messages[right].Val
			if id == vl && id == vr {
				log.Println("Node", id, "becomes the leader")
				state.Status = statusElected
				state.Leader = id
				resetMessages(&state)
				node.SendMessage(right, encodeAll(byte(msgT), id))
			} else {
				y := max(id, vl, vr)
				if y == state.Messages[left].Val {
					node.SendMessage(left, encodeAll(byte(msgA), y))
				}
				if y == state.Messages[right].Val {
					node.SendMessage(right, encodeAll(byte(msgA), y))
				}
				resetMessages(&state)
				state.Status = statusACandidate
				log.Println("Node", id, "awaits acknowledgement")
			}
		}

	case statusACandidate:
		if msg.Type == msgE { //negative acknowledgement was skipped, go passive and act so
			log.Println("Node", id, "becomes passive")
			state.Status = statusPassive
			resetMessages(&state)
			node.SendMessage(ops(sender), encodeAll(msg.Type, msg.Val))
		} else {
			if isEmptyMessage(state.Messages[sender]) {
				state.Messages[sender] = msg
			}
			if heldMessages(state.Messages) == 2 { //
				node.SendMessage(left, encodeAll(byte(msgE), id))
				node.SendMessage(right, encodeAll(byte(msgE), id))
				resetMessages(&state)
				state.Status = statusECandidate
				log.Println("Node", id, "enters next round of election")
			}
			//negative acknowledgement is not being sent,
			// so getting 2 acknowledgement messages means we are greater than our active neighbours
		}
	case statusPassive:
		if msg.Val != int64(node.GetIndex()) { //intercept messages addressed to me
			node.SendMessage(ops(sender), encodeAll(msg.Type, msg.Val))
		}
		if msg.Type == msgT {
			if sender == right {
				state.ReverseOrientation = true
			}
			state.Leader = msg.Val
			return true
		}
	case statusElected:
		if msg.Type == msgT {
			return true
		}
	}
	return false
}

func run(node lib.Node) {
	node.StartProcessing()
	initialize(node)
	var sender int
	var message []byte
	for {
		switch awaitMessage(getState(node)) {
		case left:
			sender = left
			message = node.ReceiveMessage(left)
		case right:
			sender = right
			message = node.ReceiveMessage(right)
		case both:
			sender, message = node.ReceiveAnyMessage()
		}
		if process(node, sender, message) {
			break
		}
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
	set := make(map[int]bool)
	for _, node := range nodes {
		set[node.GetIndex()] = false
	}
	node := nodes[0]
	for i := 0; i < len(nodes); i += 1 {
		if getState(node).ReverseOrientation {
			node = node.GetOutNeighbors()[left]
		} else {
			node = node.GetOutNeighbors()[right]
		}
		if set[node.GetIndex()] == true {
			panic("Ring orientation is incorrect")
		}
		set[node.GetIndex()] = true
	}
	if node.GetIndex() != nodes[0].GetIndex() {
		panic("Ring orientation is incorrect")
	}
	log.Println("Elected node", leader, "as a leader")
}
func encode(buffer *bytes.Buffer, values ...interface{}) {
	for _, value := range values {
		if binary.Write(buffer, binary.BigEndian, value) != nil {
			panic("Failed to encode a value")
		}
	}
}

func decode(buffer *bytes.Buffer, values ...interface{}) {
	for _, value := range values {
		if binary.Read(buffer, binary.BigEndian, value) != nil {
			panic("Failed to decode a value")
		}
	}
}
func encodeAll(values ...interface{}) []byte {
	buffer := bytes.Buffer{}
	encode(&buffer, values...)
	return buffer.Bytes()
}
func decodeAll(data []byte, values ...interface{}) {
	buffer := bytes.NewBuffer(data)
	decode(buffer, values...)
}
