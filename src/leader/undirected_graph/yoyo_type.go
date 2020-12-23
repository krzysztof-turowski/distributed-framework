package undirected_graph

import (
	"encoding/json"
	"lib"
	"log"
)

func toInt(b1 bool, b2 bool) int {
	value := 0
	if b1 {
		value += 2
	}
	if b2 {
		value += 1
	}
	return value
}

func toBits(value int) (int, int) {
	b1, b2 := (value & 2) >> 1, value & 1
	return b1, b2
}

type messageYoYo struct {
	Content int
	// YO-: Content is minimum value received by message sender
	// -YO: Content is 2-bit value {ab}, where a=(YES=>1; NO=>0), b=pruneRequested
}

type phaseType string

const (
	setup  phaseType = "setup"
	yoDown           = "yoDown"
	yoUp             = "yoUp"
)

type statusType int

const (
	detached statusType = 0
	source              = 1
	sink                = 2
	internal            = 3
)

func (status statusType) String() string {
	switch status {
	case source:
		return "source"
	case sink:
		return "sink"
	case internal:
		return "internal"
	case detached:
		return "detached"
	}
	return ""
}

type edgeStateType string

const (
	in     edgeStateType = "in"
	out                  = "out"
	pruned               = "pruned"
)

type stateYoYo struct {
	EdgeStates   []edgeStateType
	LastMessages []int
	Phase        phaseType
	Status       statusType
	Min          int // applicable in case Status \in {yoUp, yoDown}
}

func newStateYoYo(size int) *stateYoYo {
	return &stateYoYo{
		EdgeStates:   make([]edgeStateType, size),
		LastMessages: make([]int, size),
		Phase:        setup,
		Status:       detached,
	}
}

/*
 * Returns true if this vertex is the only left, and hence the algorithm should be terminated.
 */
func (s *stateYoYo) UpdateStatus() bool {
	hasOutEdge, hasInEdge := false, false
	for i := 0; i < len(s.EdgeStates) && !(hasOutEdge && hasInEdge); i++ {
		if s.EdgeStates[i] == out {
			hasOutEdge = true
		} else if s.EdgeStates[i] == in {
			hasInEdge = true
		}
	}

	previousStatus := s.Status
	s.Status = statusType(toInt(hasInEdge, hasOutEdge))
	isLeader := previousStatus == source && s.Status == detached
	if isLeader {
		s.Status = source
	}
	return isLeader
}

func (s *stateYoYo) FlipEdge(index int) {
	currentDirection := s.EdgeStates[index]
	if currentDirection == in {
		s.EdgeStates[index] = out
	} else if currentDirection == out {
		s.EdgeStates[index] = in
	}
}

/*
 * Returns (requestPrune []bool, pruneDefault bool)
 * If the vertex is sink and received the same value from every edge, pruneDefault is true,
 * and pruneRequest will be sent to every neighbor no matter what requestPrune[] values are.
 * Otherwise, if the vertex received the same value from *m* edges,
 * pruneRequest will be sent to *m-1* of them, for which requestPrune[i] == true.
 */
func (s *stateYoYo) PreprocessPruning(outEdges int) ([]bool, bool) {
	shouldBePruned := make([]bool, len(s.EdgeStates))
	receivedValues := make(map[int]bool)
	for i := 0; i < len(s.EdgeStates); i++ {
		if s.EdgeStates[i] == in {
			value := s.LastMessages[i]
			if receivedValues[value] {
				shouldBePruned[i] = true
			} else {
				receivedValues[value] = true
				shouldBePruned[i] = false
			}
		}
	}
	// If sink received the same value from every edge, all of them are useless
	// We can't use (s.Status == sink) instead of (outEdges == 0) since pruning preprocessing
	// takes place in the middle of the phase
	return shouldBePruned, (outEdges == 0) && (len(receivedValues) == 1)
}

func getStateYoYo(v lib.Node) *stateYoYo {
	var s stateYoYo
	json.Unmarshal(v.GetState(), &s)
	return &s
}

func setStateYoYo(v lib.Node, s *stateYoYo) {
	data, _ := json.Marshal(*s)
	v.SetState(data)
}

func sendMessageYoYo(v lib.Node, index int, message messageYoYo) {
	log.Printf("Vertex %d sends %d to edge %d\n", v.GetIndex(), message.Content, index)
	outMessage, _ := json.Marshal(message)
	v.SendMessage(index, outMessage)
}

func receiveMessageYoYo(v lib.Node, index int) messageYoYo {
	var m messageYoYo
	inMessage := v.ReceiveMessage(index)
	json.Unmarshal(inMessage, &m)
	log.Printf("Vertex %d received %d from edge %d\n", v.GetIndex(), m.Content, index)
	return m
}
