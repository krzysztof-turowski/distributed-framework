package sync_torus

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

/* NODE LABELS */
type NodeLabelType int

/* DIRECTIONS */
type DirectionType int

/* BOUNDS */
const (
	msg_bound  int           = 16
	directions DirectionType = 4
)

/* NODE TYPES */
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
	token
)

type Message struct {
	Type  MessageType
	Value NodeLabelType
	// fields to handrail
	Handrail NodeLabelType
	SoFar    int
	ToGo     int
}

/* TESTS */
const (
	testCases int = 100
)

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
	// only to check if handrail works
	State          NodeStateType
	BoundaryClosed bool
}

func newState(dim int) *StateNodeTorus {
	var st StateNodeTorus
	st.Count = 0
	st.H1 = make(map[DirectionType]NodeLabelType)
	st.H2 = make(map[DirectionType][]NodeLabelType)
	st.Pending = make([][]*Message, dim)
	for i := range st.Pending {
		st.Pending[i] = make([]*Message, 0)
	}
	// always 4 directions
	for i := DirectionType(0); i < directions; i++ {
		st.H2[i] = make([]NodeLabelType, 0)
	}
	st.State = passive
	st.BoundaryClosed = false
	return &st
}

func getState(v lib.Node) *StateNodeTorus {
	var st StateNodeTorus
	json.Unmarshal(v.GetState(), &st)
	return &st
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

func clearPendingMessages(st *StateNodeTorus) {
	for i := range st.Pending {
		st.Pending[i] = make([]*Message, 0)
	}
}

func clearAllPendingMessages(vertices []lib.Node) {
	for _, v := range vertices {
		st := getState(v)
		clearPendingMessages(st)
		setState(v, st)
	}
}

/* TORUS ORIENTATION METHODS */
func initializeNode(v lib.Node) bool {
	st := newState(v.GetOutChannelsCount())
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

	if st.Count == msg_bound {
		finish = true
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

/* BUILD */
func square_check(x int) bool {
	var int_root int = int(math.Sqrt(float64(x)))
	return (int_root * int_root) == x
}

func BuildSynchronizedTorus(n int) ([]lib.Node, lib.Synchronizer) {
	if n <= 0 {
		panic("Size cannot be <= 0.")
	}
	if !square_check(n) {
		panic("n is not a perfect square.")
	}

	if int(math.Sqrt(float64(n))) <= 4 {
		panic("Sqrt(n) should be greater than 4.")
	}

	log.Println("Building torus size", n)
	adjacencyList := make([][]int, n)

	for i := range adjacencyList {
		adjacencyList[i] = make([]int, 0)
	}

	side := int(math.Sqrt(float64(n)))

	for i := 0; i < side; i++ {
		for j := 0; j < side; j++ {
			adjacencyList[i*side+j] = append(adjacencyList[i*side+j], i*side+((j-1+side)%side))
			adjacencyList[i*side+j] = append(adjacencyList[i*side+j], i*side+((j+1+side)%side))
			adjacencyList[i*side+j] = append(adjacencyList[i*side+j], (((i-1+side)%side)*side + j))
			adjacencyList[i*side+j] = append(adjacencyList[i*side+j], (((i+1+side)%side)*side + j))
		}
	}

	return lib.BuildSynchronizedGraphFromAdjacencyList(adjacencyList, lib.GetRandomGenerator())
}

/* CHECK */
func getNodeInd(vertices []lib.Node, id NodeLabelType) int {
	for i := range vertices {
		if NodeLabelType(vertices[i].GetIndex()) == id {
			return i
		}
	}
	// not reachable
	return -1
}

func getLinkFrom(v lib.Node, id NodeLabelType) DirectionType {
	st := getState(v)
	for i, k := range st.H1 {
		if NodeLabelType(k) == id {
			return i
		}
	}
	// not reachable
	return -1
}

func checkWithoutRunning(vertices []lib.Node, initiatorNode int) {
	dMin := 1
	d := rand.Intn(len(vertices)) + dMin
	distance := 0
	currentNode := vertices[initiatorNode]
	prevNode := vertices[initiatorNode]
	st := getState(currentNode)
	k, r := findPerpendicularAny(st)
	handrail := st.H1[k]
	toGo := d
	soFar := 0
	distance++
	nextNodeId := st.H1[r]

	log.Println("Node:", currentNode.GetIndex(), "initiator.")
	log.Println("d:", d)

	for distance < 4*d {
		prevNode = currentNode
		currentNode = vertices[getNodeInd(vertices, nextNodeId)]
		soFar++
		r, k = findPerpendicularVal(getState(currentNode), handrail, getLinkFrom(currentNode, NodeLabelType(prevNode.GetIndex())))
		if soFar != toGo {
			st = getState(currentNode)
			handrail = st.H1[k]
			s := findParallelTo(st, r)
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
	numOfVer := len(vertices)
	initiatiorNode := rand.Intn(numOfVer)
	for i := 0; i < testCases; i++ {
		checkWithoutRunning(vertices, initiatiorNode)
	}

}

/* RUN */
func start(vertices []lib.Node, synchronizer lib.Synchronizer) (int, int) {
	for _, v := range vertices {
		go run(v)
	}
	synchronizer.Synchronize(2)
	synchronizer.GetStats()
	return synchronizer.GetStats()
}

func Run(n int) {
	vertices, sync := BuildSynchronizedTorus(n)
	msgs, rounds := start(vertices, sync)
	fmt.Println("ORIENTATION")
	fmt.Println("Rounds:", rounds, "Messages:", msgs)
	fmt.Println("TESTING CORECTNESS")
	multipleTests(vertices)
}

/* MY SKETCHES FOR RUNTIME BOUNDARY, MAY BE WRONG BUT USEFUL */
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
	for i := DirectionType(0); i < directions; i++ {
		for j := i + 1; j < directions; j++ {
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
	for i := DirectionType(0); i < directions; i++ {
		if emptyIntersect(st.H2[link], st.H2[i]) {
			s := i
			return s
		}
	}
	// not reachable
	return s
}

func findPerpendicularVal(st *StateNodeTorus, handrail NodeLabelType, link DirectionType) (DirectionType, DirectionType) {
	// links
	r, k := link, DirectionType(-1)
	for i := DirectionType(0); i < directions; i++ {
		if i == link {
			continue
		}
		for j := range st.H2[i] {
			if st.H2[i][j] == handrail {
				k = i
				return r, k
			}
		}

	}
	// not reachable
	return r, k
}

func markTerritory(v lib.Node, st *StateNodeTorus) {
	for i := DirectionType(0); i < DirectionType(v.GetInChannelsCount()); i++ {
		msgin := receiveMessage(v, int(i))
		if msgin != nil {
			if msgin.Type == token {
				if st.State == initiator {
					log.Println("Node:", v.GetIndex(), "boundary closed.")
					st.BoundaryClosed = true
				} else {
					sofar := msgin.SoFar + 1
					r, k := findPerpendicularVal(st, msgin.Handrail, i)
					if sofar != msgin.ToGo {
						log.Println("Node", v.GetIndex(), "straight.")
						handrail := st.H1[k]
						msgout := newMessageHandrail(NodeLabelType(v.GetIndex()), handrail, sofar, msgin.ToGo)
						s := findParallelTo(st, r)
						addPendingMessage(int(s), st, msgout)
					} else {
						log.Println("Node", v.GetIndex(), "turn.")
						handrail := st.H1[r]
						msgout := newMessageHandrail(NodeLabelType(v.GetIndex()), handrail, 0, msgin.ToGo)
						addPendingMessage(int(k), st, msgout)
					}
					st.BoundaryClosed = true
				}
			} else {
				// not reachable
			}
		}
	}
}

func mark(v lib.Node) bool {
	st := getState(v)
	finish := st.BoundaryClosed

	markTerritory(v, st)
	sendPendingMessages(v, st)
	setState(v, st)

	return finish
}

func startMarking(v lib.Node, d int) bool {
	log.Println("Node:", v.GetIndex(), "starting to mark territory.")
	st := getState(v)
	k, r := findPerpendicularAny(st)
	handrail := st.H1[k]
	msgout := newMessageHandrail(NodeLabelType(v.GetIndex()), handrail, 0, d)
	addPendingMessage(int(r), st, msgout)
	sendPendingMessages(v, st)
	setState(v, st)
	return false
}

func check(v lib.Node, d int) {
	finish := false
	st := getState(v)
	v.StartProcessing()
	if st.State == initiator {
		finish = startMarking(v, d)
	} else {
		finish = mark(v)
	}
	v.FinishProcessing(finish)
	distance := 1
	for distance < 4*d {
		v.StartProcessing()
		finish = mark(v)
		v.FinishProcessing(finish)
		distance++
	}
}

func checkAll(vertices []lib.Node, synchronizer lib.Synchronizer) (int, int) {
	clearAllPendingMessages(vertices)

	initiatorNode := 0
	dBound := int(math.Sqrt(float64(len(vertices))))
	dMin := 2
	d := rand.Intn(dBound-dMin) + dMin
	st := getState(vertices[initiatorNode])
	st.State = initiator
	setState(vertices[initiatorNode], st)

	for _, v := range vertices {
		go check(v, d)
	}
	synchronizer.Synchronize(0)
	synchronizer.GetStats()
	if !getState(vertices[initiatorNode]).BoundaryClosed {
		panic("Boundary not closed!!!")
	}
	return synchronizer.GetStats()
}
