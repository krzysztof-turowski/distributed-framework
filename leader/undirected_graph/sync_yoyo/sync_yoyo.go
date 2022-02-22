package sync_yoyo

import (
	"encoding/json"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
)

/* HELPER FUNCTIONS */

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
	return (value & 2) >> 1, value & 1
}

/* YO-YO TYPES */

type message struct {
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
	return []string{"detached", "source", "sink", "internal"}[status]
}

type edgeStateType string

const (
	in     edgeStateType = "in"
	out                  = "out"
	pruned               = "pruned"
)

type state struct {
	EdgeStates   []edgeStateType
	LastMessages []int
	Phase        phaseType
	Status       statusType
	Min          int // applicable in case Status \in {yoUp, yoDown}
}

func newState(size int) *state {
	return &state{
		EdgeStates:   make([]edgeStateType, size),
		LastMessages: make([]int, size),
		Phase:        setup,
		Status:       detached,
	}
}

/* STATE-YO-YO METHODS */

/*
 * Returns true if this vertex is the only left, and hence the algorithm should be terminated.
 */
func (s *state) UpdateStatus() bool {
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

func (s *state) FlipEdge(index int) {
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
func (s *state) PreprocessPruning(outEdges int) ([]bool, bool) {
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

/* lib.Node EXTENSION METHODS */

func getState(v lib.Node) *state {
	var s state
	json.Unmarshal(v.GetState(), &s)
	return &s
}

func setState(v lib.Node, s *state) {
	data, _ := json.Marshal(*s)
	v.SetState(data)
}

func sendMessage(v lib.Node, index int, message message) {
	log.Printf("Vertex %d sends %d to edge %d\n", v.GetIndex(), message.Content, index)
	outMessage, _ := json.Marshal(message)
	v.SendMessage(index, outMessage)
}

func receiveMessage(v lib.Node, index int) message {
	var m message
	inMessage := v.ReceiveMessage(index)
	json.Unmarshal(inMessage, &m)
	log.Printf("Vertex %d received %d from edge %d\n", v.GetIndex(), m.Content, index)
	return m
}

func receiveMessageYoDown(v lib.Node, s *state, index int) int {
	message := receiveMessage(v, index)
	s.LastMessages[index] = message.Content
	return message.Content
}

func receiveMessageYoUp(v lib.Node, s *state, index int) (bool, bool) {
	message := receiveMessage(v, index)
	vote, pruneRequested := toBits(message.Content)

	if pruneRequested == 1 {
		s.EdgeStates[index] = pruned
	}

	return vote == 1, pruneRequested == 1
}

func sendMessageYoUp(v lib.Node, index int, answer bool, pruneRequested bool) {
	outMessage := message{Content: toInt(answer, pruneRequested)}
	sendMessage(v, index, outMessage)
}

/* PHASE-PROCESSING METHODS */

func initialize(v lib.Node) bool {
	if v.GetOutChannelsCount() == 0 {
		log.Panic("Graph contains a node adjacent to 0 edges")
	}

	s := newState(v.GetOutChannelsCount())
	msg := message{Content: v.GetIndex()}
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		sendMessage(v, i, msg)
	}
	setState(v, s)
	return false
}

func processSetup(v lib.Node, s *state) {
	log.Printf("Vertex %d processing setup phase\n", v.GetIndex())
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		m := receiveMessage(v, i)
		if v.GetIndex() == m.Content {
			log.Panic("Graph contains adjacent vertices with equal ID")
		}
		b := v.GetIndex() < m.Content
		if b {
			log.Printf("Vertex %d orienting its edge %d as outgoing\n", v.GetIndex(), i)
			s.EdgeStates[i] = out
		} else {
			log.Printf("Vertex %d orienting its edge %d as ingoing\n", v.GetIndex(), i)
			s.EdgeStates[i] = in
		}
	}

	s.UpdateStatus()
	log.Printf("Vertex %d updated status is %s\n", v.GetIndex(), s.Status)
}

func processYoDown(v lib.Node, s *state) {
	log.Printf("Vertex %d (%s) processing YO- phase\n", v.GetIndex(), s.Status)
	s.Min = v.GetIndex()
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		if s.EdgeStates[i] == in {
			value := receiveMessageYoDown(v, s, i)
			if value < s.Min {
				s.Min = value
			}
		}
	}

	outMessage := message{Content: s.Min}
	log.Printf("Vertex %d (%s) passing its minimum %d to neighbors\n", v.GetIndex(), s.Status, s.Min)
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		if s.EdgeStates[i] == out {
			sendMessage(v, i, outMessage)
		}
	}
}

func processYoUp(v lib.Node, s *state) bool {
	log.Printf("Vertex %d (%s) processing -YO phase\n", v.GetIndex(), s.Status)
	finalVote := true
	needsFlipping := make([]bool, v.GetOutChannelsCount())
	outEdgesLeft := 0
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		if s.EdgeStates[i] == out {
			vote, pruned := receiveMessageYoUp(v, s, i)
			finalVote = finalVote && vote
			needsFlipping[i] = !vote
			if !pruned {
				outEdgesLeft++
			}
		}
	}

	requestPrune, pruneDefault := s.PreprocessPruning(outEdgesLeft)
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		if s.EdgeStates[i] == in {
			vote := finalVote && (s.Min == s.LastMessages[i])
			if pruneDefault || requestPrune[i] {
				log.Printf("Vertex %d (%s) at edge %d votes %t and request pruning\n", v.GetIndex(), s.Status, i, vote)
			} else {
				log.Printf("Vertex %d (%s) at edge %d votes %t\n", v.GetIndex(), s.Status, i, vote)
			}
			sendMessageYoUp(v, i, vote, pruneDefault || requestPrune[i])
			needsFlipping[i] = !vote
			if pruneDefault || requestPrune[i] {
				s.EdgeStates[i] = pruned
			}
		}
	}

	for i := 0; i < v.GetOutChannelsCount(); i++ {
		if needsFlipping[i] {
			s.FlipEdge(i)
		}
	}

	isLeader := s.UpdateStatus()
	log.Printf("Vertex %d updated status is %s\n", v.GetIndex(), s.Status)
	return isLeader
}

/* MAIN PART */

func process(v lib.Node, round int) bool {
	s, finish := getState(v), false
	switch s.Phase {
	case setup:
		processSetup(v, s)
		s.Phase = yoDown
	case yoDown:
		processYoDown(v, s)
		s.Phase = yoUp
	case yoUp:
		finish = processYoUp(v, s)
		s.Phase = yoDown
	}
	setState(v, s)
	return finish || s.Status == detached
}

func run(v lib.Node) {
	v.StartProcessing()
	finish := initialize(v)
	v.FinishProcessing(finish)

	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = process(v, round)
		v.FinishProcessing(finish)
	}
}

func check(vertices []lib.Node) {
	min := vertices[0].GetIndex()
	leader := -1
	for _, v := range vertices {
		s := getState(v)
		if s.Status != detached {
			log.Printf("Leader is %d\n", v.GetIndex())
			leader = v.GetIndex()
		}
		if min > v.GetIndex() {
			min = v.GetIndex()
		}
	}
	log.Printf("Minimum ID was %d\n", min)
	if min != leader {
		panic("Algorithm's output is wrong")
	}
}

/* ENTRY POINTS */

func Run(vertices []lib.Node, synchronizer lib.Synchronizer) {
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go run(v)
	}
	synchronizer.Synchronize(0)
	synchronizer.GetStats()
	check(vertices)
}

func RunRandom(n int, p float64) {
	Run(lib.BuildSynchronizedRandomGraph(n, p))
}
