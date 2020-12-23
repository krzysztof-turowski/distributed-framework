package undirected_graph

import (
	"encoding/json"
	"lib"
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
	return []string{"detached", "source", "sink", "internal"}[status]
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

/* STATE-YO-YO METHODS */

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

/* lib.Node EXTENSION METHODS */

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

func receiveMessageYoDown(v lib.Node, s *stateYoYo, index int) int {
	message := receiveMessageYoYo(v, index)
	s.LastMessages[index] = message.Content
	return message.Content
}

func receiveMessageYoUp(v lib.Node, s *stateYoYo, index int) (bool, bool) {
	message := receiveMessageYoYo(v, index)
	vote, pruneRequested := toBits(message.Content)

	if pruneRequested == 1 {
		s.EdgeStates[index] = pruned
	}

	return vote == 1, pruneRequested == 1
}

func sendMessageYoUp(v lib.Node, index int, answer bool, pruneRequested bool) {
	outMessage := messageYoYo{Content: toInt(answer, pruneRequested)}
	sendMessageYoYo(v, index, outMessage)
}

/* PHASE-PROCESSING METHODS */

func initializeYoYo(v lib.Node) bool {
	if v.GetOutChannelsCount() == 0 {
		log.Panic("Graph contains a node adjacent to 0 edges")
	}

	s := newStateYoYo(v.GetOutChannelsCount())
	msg := messageYoYo{Content: v.GetIndex()}
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		sendMessageYoYo(v, i, msg)
	}
	setStateYoYo(v, s)
	return false
}

func processSetup(v lib.Node, s *stateYoYo) {
	log.Printf("Vertex %d processing setup phase\n", v.GetIndex())
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		m := receiveMessageYoYo(v, i)
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

func processYoDown(v lib.Node, s *stateYoYo) {
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

	outMessage := messageYoYo{Content: s.Min}
	log.Printf("Vertex %d (%s) passing its minimum %d to neighbors\n", v.GetIndex(), s.Status, s.Min)
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		if s.EdgeStates[i] == out {
			sendMessageYoYo(v, i, outMessage)
		}
	}
}

func processYoUp(v lib.Node, s *stateYoYo) bool {
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

func processYoYo(v lib.Node, round int) bool {
	s, finish := getStateYoYo(v), false
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
	setStateYoYo(v, s)
	return finish || s.Status == detached
}

func runYoYo(v lib.Node) {
	v.StartProcessing()
	finish := initializeYoYo(v)
	v.FinishProcessing(finish)

	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = processYoYo(v, round)
		v.FinishProcessing(finish)
	}
}

func checkYoYo(vertices []lib.Node) {
	min := vertices[0].GetIndex()
	leader := -1
	for _, v := range vertices {
		s := getStateYoYo(v)
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

func RunYoYo(vertices []lib.Node, synchronizer lib.Synchronizer) {
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runYoYo(v)
	}
	synchronizer.Synchronize(0)
	synchronizer.GetStats()
	checkYoYo(vertices)
}

func RunYoYoRandom(n int, p float64) {
	RunYoYo(lib.BuildSynchronizedRandomGraph(n, p))
}
