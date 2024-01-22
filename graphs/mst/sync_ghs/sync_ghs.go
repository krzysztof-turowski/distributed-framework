package sync_ghs

import (
	"encoding/json"
	"fmt"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
	"reflect"
	"sort"
)

//	EDGES
//
// Synchronized GHS algorithm for finding Minimal Spanning Tree requires every edge to have a unique weight.
// To achieve that for every u,v in E we define new edge weight as a triple (w, v, u) where w is the original
// weight and v < u. We order such triplets lexicographically.
type edge struct {
	Weight   int
	SmallerV int
	LargerV  int
}

func newEdge(weight, u, v int) *edge {
	if u > v {
		u, v = v, u
	}
	return &edge{Weight: weight, SmallerV: u, LargerV: v}
}

func (e *edge) Less(otherEdge lib.Comparable) bool {
	f := otherEdge.(*edge)
	if e.Weight != f.Weight {
		return e.Weight < f.Weight
	}
	if e.SmallerV != f.SmallerV {
		return e.SmallerV < f.SmallerV
	}
	return e.LargerV < f.LargerV
}

func (e *edge) isConnected(v int) bool {
	return v == e.SmallerV || v == e.LargerV
}

type edgeStatus int

const (
	outgoingEdge edgeStatus = iota
	treeEdge
	rejectedEdge
)

//  MESSAGES
// Messages used in synchronized ghs algorithm

type messageType int

const (
	nilMessage messageType = iota
	messageVerifyEdge
	messageProposeEdge
	messageChooseEdge
	messsageConnectEdge
	messageElectNewRoot
	messageCompleted
)

type messageSynchronizedGHS struct {
	Type  messageType
	MWOE  *edge // MWOE - Minimal Weight Outgoing Edge
	Index int
	Root  int
}

func sendMessageSynchronizedGHS(v lib.WeightedGraphNode, index int, message *messageSynchronizedGHS) {
	outMessage, _ := json.Marshal(message)
	v.SendMessage(index, outMessage)
}

func receiveMessageSynchronizedGHS(v lib.WeightedGraphNode, index int) *messageSynchronizedGHS {
	var inMessage messageSynchronizedGHS
	if data := v.ReceiveMessage(index); data != nil {
		if err := json.Unmarshal(data, &inMessage); err == nil {
			return &inMessage
		}
	}
	return nil
}

/*  Vertice state in Synchronized GHS algorithm
 * State contains statuses for every edge and information about the current tree.
 */

type stateSynchronizedGHSType int

const (
	stateVerifyOutgoingEdges stateSynchronizedGHSType = iota
	statePropagateUpMWOE
	statePropagateDownChosenMWOE
	stateElectNewRoot
	stateBroadcastNewRoot
	nilParent int = -1
)

type stateSynchronizedGHS struct {
	Status            stateSynchronizedGHSType
	BroadcastRound    int
	Edges             []edgeStatus
	TreeRoot          int
	TreeParent        int
	ProposedEdge      *edge
	ProposedEdgeIndex int
	SelectedEdge      bool
}

// State encoding functions:

func readState(v lib.WeightedGraphNode) *stateSynchronizedGHS {
	var state stateSynchronizedGHS
	data := v.GetState()
	if err := json.Unmarshal(data, &state); err == nil {
		return &state
	}
	return nil
}

func (s *stateSynchronizedGHS) saveState(v lib.WeightedGraphNode) {
	data, _ := json.Marshal(s)
	v.SetState(data)
}

// Helper functions for iterating over children/tree edges/outgoing edges

func (s *stateSynchronizedGHS) foreachChildEdge(f func(i int)) {
	for i, eStatus := range s.Edges {
		if eStatus == treeEdge && i != s.TreeParent {
			f(i)
		}
	}
}

func (s *stateSynchronizedGHS) foreachOutgoingEdge(f func(i int)) {
	for i, eStatus := range s.Edges {
		if eStatus == outgoingEdge {
			f(i)
		}
	}
}

func (s *stateSynchronizedGHS) foreachTreeEdge(f func(i int)) {
	for i, eStatus := range s.Edges {
		if eStatus == treeEdge {
			f(i)
		}
	}
}

// functions for processing round

func (s *stateSynchronizedGHS) sendVerifyEdges(v lib.WeightedGraphNode) {
	msgTest := &messageSynchronizedGHS{Type: messageVerifyEdge, Index: v.GetIndex(), Root: s.TreeRoot}
	s.foreachOutgoingEdge(func(i int) {
		sendMessageSynchronizedGHS(v, i, msgTest)
	})
}

func (s *stateSynchronizedGHS) proposeMWOE(v lib.WeightedGraphNode) *edge {
	s.ProposedEdge = nil
	s.ProposedEdgeIndex = -1
	s.SelectedEdge = false
	s.foreachOutgoingEdge(func(i int) {
		message := receiveMessageSynchronizedGHS(v, i)
		e := newEdge(v.GetInWeights()[i], v.GetIndex(), message.Index)
		if s.TreeRoot == message.Root {
			s.Edges[i] = rejectedEdge
		} else if s.ProposedEdge == nil || e.Less(s.ProposedEdge) {
			s.ProposedEdge = e
			s.ProposedEdgeIndex = i
		}
	})
	return s.ProposedEdge
}

func (s *stateSynchronizedGHS) propagateUpProposedMWOE(v lib.WeightedGraphNode, e *edge) {
	if s.TreeParent != nilParent {
		if e != nil {
			sendMessageSynchronizedGHS(v, s.TreeParent, &messageSynchronizedGHS{Type: messageProposeEdge, MWOE: e})
		} else {
			v.SendMessage(s.TreeParent, nil)
		}
	}
}

func (s *stateSynchronizedGHS) receiveProposedMWOE(v lib.WeightedGraphNode) *edge {
	foundBetterProposedEdge := false
	s.foreachChildEdge(func(i int) {
		message := receiveMessageSynchronizedGHS(v, i)
		if message != nil && (s.ProposedEdge == nil || message.MWOE.Less(s.ProposedEdge)) {
			s.ProposedEdge = message.MWOE
			s.ProposedEdgeIndex = i
			foundBetterProposedEdge = true
		}
	})
	if foundBetterProposedEdge {
		return s.ProposedEdge
	}
	return nil
}

func (s *stateSynchronizedGHS) choseMWOE(v lib.WeightedGraphNode) messageType {
	if s.TreeRoot == v.GetIndex() {
		if s.ProposedEdge != nil {
			s.SelectedEdge = s.ProposedEdge.isConnected(v.GetIndex())
			return messageChooseEdge
		} else {
			return messageCompleted
		}
	}
	return nilMessage
}

func (s *stateSynchronizedGHS) propagateDownChosenMWOE(v lib.WeightedGraphNode, message messageType) {
	switch message {
	case messageChooseEdge:
		s.foreachChildEdge(func(i int) {
			if i == s.ProposedEdgeIndex {
				sendMessageSynchronizedGHS(v, s.ProposedEdgeIndex, &messageSynchronizedGHS{Type: messageChooseEdge})
			} else {
				v.SendMessage(i, nil)
			}
		})
	case messageCompleted:
		s.foreachChildEdge(func(i int) {
			sendMessageSynchronizedGHS(v, i, &messageSynchronizedGHS{Type: messageCompleted})
		})
	default:
		s.foreachChildEdge(func(i int) {
			v.SendMessage(i, nil)
		})
	}
}

func (s *stateSynchronizedGHS) receiveChosenMWOE(v lib.WeightedGraphNode) messageType {
	if s.TreeParent != nilParent {
		message := receiveMessageSynchronizedGHS(v, s.TreeParent)
		if message != nil && message.Type == messageChooseEdge {
			s.SelectedEdge = s.ProposedEdge.isConnected(v.GetIndex())
		}
		if message != nil {
			return message.Type
		}
	}
	return nilMessage
}

func (s *stateSynchronizedGHS) sendConnectComponents(v lib.WeightedGraphNode) {
	s.foreachOutgoingEdge(func(i int) {
		if s.SelectedEdge && i == s.ProposedEdgeIndex {
			s.Edges[i] = treeEdge
			sendMessageSynchronizedGHS(v, i, &messageSynchronizedGHS{Type: messsageConnectEdge})
			log.Printf("Adding edge (%d, %d) with weight %d to MST\n", s.ProposedEdge.LargerV, s.ProposedEdge.SmallerV, s.ProposedEdge.Weight)
		} else {
			v.SendMessage(i, nil)
		}
	})
}

func (s *stateSynchronizedGHS) electNewRoot(v lib.WeightedGraphNode) *int {
	var newRoot *int = nil
	if s.SelectedEdge {
		msg := receiveMessageSynchronizedGHS(v, s.ProposedEdgeIndex)
		if msg != nil && msg.Type == messsageConnectEdge && v.GetIndex() == s.ProposedEdge.LargerV {
			s.TreeParent = nilParent
			s.TreeRoot = v.GetIndex()
			newRoot = &s.TreeRoot
		}
	}
	s.foreachOutgoingEdge(func(i int) {
		msg := receiveMessageSynchronizedGHS(v, i)
		if msg != nil && msg.Type == messsageConnectEdge {
			s.Edges[i] = treeEdge
		}
	})
	return newRoot
}

func (s *stateSynchronizedGHS) propagateDownNewRoot(v lib.WeightedGraphNode, newRoot *int) {
	s.foreachTreeEdge(func(i int) {
		if newRoot != nil && i != s.TreeParent {
			sendMessageSynchronizedGHS(v, i, &messageSynchronizedGHS{Type: messageElectNewRoot, Root: *newRoot})
		} else {
			v.SendMessage(i, nil)
		}
	})
}

func (s *stateSynchronizedGHS) receiveNewRoot(v lib.WeightedGraphNode) *int {
	var newRoot *int = nil
	s.foreachTreeEdge(func(i int) {
		message := receiveMessageSynchronizedGHS(v, i)
		if message != nil {
			s.TreeRoot = message.Root
			s.TreeParent = i
			newRoot = &message.Root
		}
	})
	return newRoot
}

// Initialize state and send verification messages for outgoing edges
func initializeSynchronizedGHS(v lib.WeightedGraphNode) bool {
	s := &stateSynchronizedGHS{
		TreeRoot:          v.GetIndex(),
		Edges:             make([]edgeStatus, len(v.GetOutNeighbors())),
		TreeParent:        nilParent,
		ProposedEdge:      nil,
		ProposedEdgeIndex: -1,
		SelectedEdge:      false,
	}
	for i := range v.GetOutNeighbors() {
		s.Edges[i] = outgoingEdge
	}
	s.sendVerifyEdges(v)
	s.Status = stateVerifyOutgoingEdges
	s.saveState(v)
	return false
}

// The algorithm requires each node to know the number of nodes in the network
func processSynchronizedGHS(v lib.WeightedGraphNode, n int) bool {
	s := readState(v)
	finished := false
	switch s.Status {
	// Verification of outgoing edges - for every non-tree edge check if it has different leader(root).
	// Reject edges pointing to vertices with the same leader. Choose MWOE (minimal weight outgoing edge)
	case stateVerifyOutgoingEdges:
		mwoe := s.proposeMWOE(v)
		s.propagateUpProposedMWOE(v, mwoe)
		s.Status = statePropagateUpMWOE
		s.BroadcastRound = 0

	// Sending proposed MWOE to parent. After n rounds the root has MWOE for the whole tree so
	// the root chooses MWOE and sends information to connect to the vertice which initially proposed this MWOE.
	// If there is no MWOE then MST is completed and procedure is finished.
	case statePropagateUpMWOE:
		mwoe := s.receiveProposedMWOE(v)
		s.BroadcastRound++
		if s.BroadcastRound < n {
			s.propagateUpProposedMWOE(v, mwoe)
		} else {
			msg := s.choseMWOE(v)
			s.propagateDownChosenMWOE(v, msg)
			s.Status = statePropagateDownChosenMWOE
			s.BroadcastRound = 0
			finished = msg == messageCompleted
		}

	// Sending elected node to original proposer or broadcasting information about completed MST.
	case statePropagateDownChosenMWOE:
		msg := s.receiveChosenMWOE(v)
		s.BroadcastRound++
		if s.BroadcastRound < n {
			s.propagateDownChosenMWOE(v, msg)
		} else {
			s.sendConnectComponents(v)
			s.Status = stateElectNewRoot
		}
		finished = msg == messageCompleted

	// Electing new root. For every new tree there is a single edge e which was a MWOE for
	// two trees in the previous round. The vertice with the larger index connected to e becomes new root.
	case stateElectNewRoot:
		root := s.electNewRoot(v)
		s.propagateDownNewRoot(v, root)
		s.Status = stateBroadcastNewRoot
		s.BroadcastRound = 0

	// Broadcasting information about new root. After n rounds all vertices in tree have received it.
	// Then verify every outgoing edge if it points to a different tree.
	case stateBroadcastNewRoot:
		root := s.receiveNewRoot(v)
		s.BroadcastRound++
		if s.BroadcastRound < n {
			s.propagateDownNewRoot(v, root)
		} else {
			s.sendVerifyEdges(v)
			s.Status = stateVerifyOutgoingEdges
		}
	}
	s.saveState(v)
	return finished
}

func runSynchronizedGHS(v lib.WeightedGraphNode, n int) {
	v.StartProcessing()
	finish := initializeSynchronizedGHS(v)
	v.FinishProcessing(finish)

	for !finish {
		v.StartProcessing()
		finish = processSynchronizedGHS(v, n)
		v.FinishProcessing(finish)
	}
}

func RunSynchronizedGHS(vertices []lib.WeightedGraphNode, synchronizer lib.Synchronizer) {
	for _, v := range vertices {
		go runSynchronizedGHS(v, len(vertices))
	}
	synchronizer.Synchronize(0)
	synchronizer.GetStats()
	verifySynchronizedGHS(vertices)
}

func RunSynchronizedGHSRandom(n, m, maxWeight int) {
	vertices, synchronizer := lib.BuildSynchronizedRandomConnectedWeightedGraph(n, m, maxWeight, lib.GetGenerator())
	RunSynchronizedGHS(vertices, synchronizer)
}

//	VERIFICATION
//
// assuming all edges have unique weights than there is single MST
func verifySynchronizedGHS(vertices []lib.WeightedGraphNode) {
	expected := findMST(vertices)
	for i, v := range vertices {
		actual := getTreeEdges(v)
		sort.Ints(actual)
		sort.Ints(expected[i])
		if !reflect.DeepEqual(actual, expected[i]) {
			panic(fmt.Sprintf("Node %d has tree neighbors %v but should have %v", v.GetIndex(), actual, expected[i]))
		}
	}
	log.Printf("Algorithm output is correct\n")
}

func getTreeEdges(v lib.WeightedGraphNode) []int {
	tree := make([]int, 0)
	for i, status := range readState(v).Edges {
		if status == treeEdge {
			tree = append(tree, v.GetOutNeighbors()[i].GetIndex())
		}
	}
	return tree
}
