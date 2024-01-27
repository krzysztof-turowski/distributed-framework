package sync_torus

import (
	"encoding/json"
	"fmt"
	"log"
	"math"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

/* TESTS */
const (
	TEST_CASES int = 10
)

/* NODE LABELS */
type NodeLabelType int

/* DIRECTIONS */
type DirectionType int

type SenseOfDirectionMaps struct {
	Processed                 bool
	ParallelLink              map[DirectionType]DirectionType
	PerpendicularWithHandrail map[DirectionType](map[NodeLabelType]DirectionType)
}

func newSenseOfDirectionMaps() *SenseOfDirectionMaps {
	var m SenseOfDirectionMaps
	m.Processed = false
	m.ParallelLink = make(map[DirectionType]DirectionType)
	m.PerpendicularWithHandrail = make(map[DirectionType](map[NodeLabelType]DirectionType))
	for i := DirectionType(0); i < DIRECTIONS; i++ {
		m.PerpendicularWithHandrail[i] = make(map[NodeLabelType]DirectionType)
	}
	return &m
}

func processInformation(st *StateNodeTorus) {
	if !getDirections(st).Processed {
		// parallel
		for i := DirectionType(0); i < DIRECTIONS; i++ {
			getDirections(st).ParallelLink[i] = findParallelTo(st, i)
		}
		// perpendicular including handrail info
		for i := DirectionType(0); i < DIRECTIONS; i++ {
			for j := range st.H2[i] {
				handrail := st.H2[i][j]
				link := findPerpendicularTo(st, handrail, i)
				if link == -1 {
					// parallel link
					continue
				}
				getDirections(st).PerpendicularWithHandrail[i][handrail] = link
			}
		}
		getDirections(st).Processed = true
	}
}

/* BOUNDS */
const (
	MSG_BOUND  int           = 16
	DIRECTIONS DirectionType = 4
)

/* NODE TYPES ONLY USED IN SKETCHES */
type NodeStateType int

const (
	initiator NodeStateType = iota
	passive
)

/* MESSAGE TYPES */
type MessageType int

const (
	_ MessageType = iota
	one
	two
	token // for leader election
)

type Message struct {
	Type  MessageType
	Value NodeLabelType
	// fields to handrail
	Handrail NodeLabelType
	SoFar    int
	ToGo     int
}

/* STATE FUNCTIONS */
type StateNodeTorus struct {
	// counter of received messages
	Count int
	// H1 map
	H1 map[DirectionType]NodeLabelType
	// H2 map
	H2 map[DirectionType][]NodeLabelType
	// messages waiting to be sent (one per neighbor per round)
	Pending [][]*Message
	// sense of direction
	Directions *SenseOfDirectionMaps
}

func newState() *StateNodeTorus {
	st := StateNodeTorus{
		Count:      0,
		H1:         make(map[DirectionType]NodeLabelType),
		H2:         make(map[DirectionType][]NodeLabelType),
		Pending:    make([][]*Message, DIRECTIONS),
		Directions: newSenseOfDirectionMaps(),
	}
	for i := range st.Pending {
		st.Pending[i] = make([]*Message, 0)
	}
	// always 4 directions
	for i := DirectionType(0); i < DIRECTIONS; i++ {
		st.H2[i] = make([]NodeLabelType, 0)
	}
	return &st
}

func getState(v lib.Node) *StateNodeTorus {
	var st StateNodeTorus
	json.Unmarshal(v.GetState(), &st)
	return &st
}

func getDirections(st *StateNodeTorus) *SenseOfDirectionMaps {
	return st.Directions
}

func setState(v lib.Node, st *StateNodeTorus) {
	data, _ := json.Marshal(*st)
	v.SetState(data)
}

/* MESSAGE FUNCTIONS */
func newMessageOrient(msgtype MessageType, value NodeLabelType) *Message {
	msg := Message{
		Type:     msgtype,
		Value:    value,
		Handrail: -1,
		SoFar:    -1,
		ToGo:     -1,
	}
	return &msg
}

func newMessageHandrail(value NodeLabelType, handrail NodeLabelType, sofar int, togo int) *Message {
	msg := Message{
		Type:     token,
		Value:    value,
		Handrail: handrail,
		SoFar:    sofar,
		ToGo:     togo,
	}
	return &msg
}

func sendMessage(v lib.Node, index int, msg *Message) {
	if msg != nil {
		out, _ := json.Marshal(*msg)
		v.SendMessage(index, out)
	} else {
		v.SendMessage(index, nil)
	}
}

func receiveMessage(v lib.Node, index int) *Message {
	var msg Message
	in := v.ReceiveMessage(index)

	if in != nil {
		json.Unmarshal(in, &msg)
		return &msg
	}
	return nil
}

func sendPendingMessages(v lib.Node, st *StateNodeTorus) {
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		if len(st.Pending[i]) > 0 {
			sendMessage(v, i, st.Pending[i][0])
			st.Pending[i] = st.Pending[i][1:]
		} else {
			sendMessage(v, i, nil)
		}
	}
}

func addPendingMessage(index int, st *StateNodeTorus, msg *Message) {
	st.Pending[index] = append(st.Pending[index], msg)
}

/* TORUS ORIENTATION METHODS */
func initializeNode(v lib.Node) bool {
	st := newState()
	log.Println("Node", v.GetIndex(), "started orientation protocol")

	msg := newMessageOrient(one, NodeLabelType(v.GetIndex()))
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		addPendingMessage(i, st, msg)
	}
	sendPendingMessages(v, st)
	setState(v, st)

	return false
}

func contains(dir DirectionType, val NodeLabelType, st *StateNodeTorus) bool {
	if len(st.H2[dir]) == 0 {
		return false
	}
	for v := range st.H2[dir] {
		if NodeLabelType(st.H2[dir][v]) == val {
			return true
		}
	}
	return false
}

func process(v lib.Node, st *StateNodeTorus) {
	for i := DirectionType(0); i < DirectionType(v.GetInChannelsCount()); i++ {
		msgin := receiveMessage(v, int(i))
		if msgin != nil {
			if msgin.Type == one {
				log.Println("Node", v.GetIndex(), "orientation protocol phase 1")
				st.H1[i] = msgin.Value
				id2 := msgin.Value
				msgout := newMessageOrient(two, id2)
				for j := DirectionType(0); j < DirectionType(v.GetOutChannelsCount()); j++ {
					if j != i {
						addPendingMessage(int(j), st, msgout)
					}
				}
				st.Count++
			} else if msgin.Type == two {
				log.Println("Node", v.GetIndex(), "orientation protocol phase 2")
				id2 := msgin.Value
				if !contains(i, id2, st) {
					st.H2[i] = append(st.H2[i], id2)
				}
				st.Count++
			}
		}
	}
}

func orient(v lib.Node) bool {
	st := getState(v)
	finish := false

	if st.Count == MSG_BOUND {
		finish = true
		// process links
		processInformation(st)
	}

	process(v, st)
	sendPendingMessages(v, st)
	setState(v, st)

	return finish
}

func run(v lib.Node) {
	v.StartProcessing()
	finish := initializeNode(v)
	v.FinishProcessing(finish)

	for !finish {
		v.StartProcessing()
		finish = orient(v)
		v.FinishProcessing(finish)
	}
	log.Println("Node", v.GetIndex(), "finished.")
}

/* CHECK */
func getNodeInd(vertices []lib.Node, id NodeLabelType, prevInd int) int {
	side := int(math.Sqrt(float64(len(vertices))))
	row := int(math.Floor(float64(prevInd) / float64(side)))
	offset := prevInd - row*side

	// indexes to check
	nodesToCheck := make([]int, 0)

	// -1 row
	if row-1 < 0 {
		nodesToCheck = append(nodesToCheck, (side-1)*side+offset)
	} else {
		nodesToCheck = append(nodesToCheck, (row-1)*side+offset)
	}
	// +1 row
	if row+1 > side-1 {
		nodesToCheck = append(nodesToCheck, 0+offset)
	} else {
		nodesToCheck = append(nodesToCheck, (row+1)*side+offset)
	}
	// -1 column
	if offset-1 < 0 {
		nodesToCheck = append(nodesToCheck, row*side+side-1)
	} else {
		nodesToCheck = append(nodesToCheck, row*side+offset-1)
	}
	// +1 column
	if offset+1 > side-1 {
		nodesToCheck = append(nodesToCheck, row*side)
	} else {
		nodesToCheck = append(nodesToCheck, row*side+offset+1)
	}

	for i := range nodesToCheck {
		if NodeLabelType(vertices[nodesToCheck[i]].GetIndex()) == id {
			return nodesToCheck[i]
		}
	}
	// not reachable
	return -1
}

func getLinkFromId(v lib.Node, id NodeLabelType) DirectionType {
	st := getState(v)
	for i, k := range st.H1 {
		if NodeLabelType(k) == id {
			return i
		}
	}
	// not reachable
	return -1
}

func emptyIntersect(h21 []NodeLabelType, h22 []NodeLabelType) bool {
	if len(h22) == 0 || len(h21) == 0 {
		return true
	}
	for i := range h21 {
		for j := range h22 {
			if h21[i] == h22[j] {
				return false
			}
		}
	}
	return true
}

func findPerpendicularAny(st *StateNodeTorus) (DirectionType, DirectionType) {
	// links
	k, r := DirectionType(-1), DirectionType(-1)
	for i := DirectionType(0); i < DIRECTIONS; i++ {
		for j := i + 1; j < DIRECTIONS; j++ {
			if !emptyIntersect(st.H2[i], st.H2[j]) {
				k, r = i, j
				break
			}
		}
		if k != -1 && r != -1 {
			break
		}
	}
	return k, r
}

func findParallelTo(st *StateNodeTorus, link DirectionType) DirectionType {
	// links
	s := DirectionType(-1)
	for i := DirectionType(0); i < DIRECTIONS; i++ {
		if emptyIntersect(st.H2[link], st.H2[i]) {
			s := i
			return s
		}
	}
	// not reachable
	return s
}

func findPerpendicularTo(st *StateNodeTorus, handrail NodeLabelType, link DirectionType) DirectionType {
	// links
	k := DirectionType(-1)
	for i := DirectionType(0); i < DIRECTIONS; i++ {
		if i == link {
			continue
		}
		for j := range st.H2[i] {
			if st.H2[i][j] == handrail {
				k = i
				return k
			}
		}

	}
	// reachable only if given parallel handrail
	return k
}

func checkWithoutRunning(vertices []lib.Node, initiatorNode int) {
	gen := lib.GetRandomGenerator()
	dMin := 1
	d := gen.Int()%len(vertices) + dMin
	distance := 0
	currentNode := vertices[initiatorNode]
	currentNodeInd := initiatorNode
	prevNode := vertices[initiatorNode]
	st := getState(currentNode)

	// this is necessary, you have to specify rotation
	k, r := findPerpendicularAny(st)
	handrail := st.H1[k]
	toGo := d
	soFar := 0
	nextNodeId := st.H1[r]

	log.Println("Node:", currentNode.GetIndex(), "initiator.")
	log.Println("d:", d)

	distance++

	for distance < 4*d {
		prevNode = currentNode
		currentNodeInd = getNodeInd(vertices, nextNodeId, currentNodeInd)
		currentNode = vertices[currentNodeInd]
		st = getState(currentNode)
		soFar++
		r = getLinkFromId(currentNode, NodeLabelType(prevNode.GetIndex()))
		k = getDirections(st).PerpendicularWithHandrail[r][handrail]
		if soFar != toGo {
			st = getState(currentNode)
			handrail = st.H1[k]
			s := getDirections(st).ParallelLink[r]
			nextNodeId = st.H1[s]
		} else {
			st = getState(currentNode)
			handrail = st.H1[r]
			soFar = 0
			nextNodeId = st.H1[k]
		}
		log.Println("Next node", nextNodeId, ".")
		distance++
	}

	if nextNodeId != NodeLabelType(vertices[initiatorNode].GetIndex()) {
		panic("Boundary not closed!!!")
	}
}

func multipleTests(vertices []lib.Node) {
	gen := lib.GetRandomGenerator()
	initiatiorNode := gen.Int() % len(vertices)
	for i := 0; i < TEST_CASES; i++ {
		checkWithoutRunning(vertices, initiatiorNode)
	}

}

/* RUN */
func start(vertices []lib.Node, synchronizer lib.Synchronizer) (int, int) {
	for _, v := range vertices {
		go run(v)
	}
	synchronizer.Synchronize(0)
	return synchronizer.GetStats()
}

func Run(n int) {
	vertices, sync := lib.BuildSynchronizedTorus(n)
	msgs, rounds := start(vertices, sync)
	fmt.Println("ORIENTATION")
	fmt.Println("Rounds:", rounds, "Messages:", msgs)
	fmt.Println("TESTING CORECTNESS")
	multipleTests(vertices)
}
