package mst

import (
	"encoding/json"
	"lib"
	"log"
)

/*  Vertice state in Synchronized GHS algorithm
 * State contain statuses for every edge adn information about the current tree.
 */

type stateSynchGHSType int

const (
	stateVerifyOutgoingEdges     stateSynchGHSType = 0
	statePropagateUpMWOE         stateSynchGHSType = 1
	statePropagateDownChosenMWOE stateSynchGHSType = 2
	stateElectNewRoot            stateSynchGHSType = 3
	stateBroadcaseNewRoot        stateSynchGHSType = 4
	nilParent                    int               = -1
)

type stateSynchGHS struct {
	Status            stateSynchGHSType
	BroadcastRound    int
	Edges             []edgeStatus
	TreeRoot          int
	TreeParent        int
	ProposedEdge      *edge
	ProposedEdgeIndex int
	SelectedEdge      bool
}

// State encoding functions:

func readState(v lib.WeightedGraphNode) *stateSynchGHS {
	var state stateSynchGHS
	data := v.GetState()
	if err := json.Unmarshal(data, &state); err == nil {
		return &state
	}
	return nil
}

func (s *stateSynchGHS) saveState(v lib.WeightedGraphNode) {
	data, _ := json.Marshal(s)
	v.SetState(data)
}

// Helper functions for iterating over children/tree edges/outgoing edges

func (s *stateSynchGHS) foreachChildEdge(f func(i int)) {
	for i, eStatus := range s.Edges {
		if eStatus == treeEdge && i != s.TreeParent {
			f(i)
		}
	}
}

func (s *stateSynchGHS) foreachOutgoingEdge(f func(i int)) {
	for i, eStatus := range s.Edges {
		if eStatus == outgoingEdge {
			f(i)
		}
	}
}

func (s *stateSynchGHS) foreachTreeEdge(f func(i int)) {
	for i, eStatus := range s.Edges {
		if eStatus == treeEdge {
			f(i)
		}
	}
}

// functions for processing round

func (s *stateSynchGHS) sendVerifyEdges(v lib.WeightedGraphNode) {
	msgTest := &messageSynchGHS{Type: msgVerifyEdge, Index: v.GetIndex(), Root: s.TreeRoot}
	s.foreachOutgoingEdge(func(i int) {
		sendMessageSynchGHS(v, i, msgTest)
	})
}

func (s *stateSynchGHS) proposeMWOE(v lib.WeightedGraphNode) *edge {
	s.ProposedEdge = nil
	s.ProposedEdgeIndex = -1
	s.SelectedEdge = false
	s.foreachOutgoingEdge(func(i int) {
		message := receiveMessageSynchGHS(v, i)
		e := newEdge(v.GetInWeights()[i], v.GetIndex(), message.Index)
		if s.TreeRoot == message.Root {
			s.Edges[i] = rejectedEdge
		} else if s.ProposedEdge == nil || e.isLess(s.ProposedEdge) {
			s.ProposedEdge = e
			s.ProposedEdgeIndex = i
		}
	})
	return s.ProposedEdge
}

func (s *stateSynchGHS) propagateUpProposedMWOE(v lib.WeightedGraphNode, e *edge) {
	if s.TreeParent != nilParent {
		if e != nil {
			sendMessageSynchGHS(v, s.TreeParent, &messageSynchGHS{Type: msgProposeEdge, MWOE: e})
		} else {
			v.SendMessage(s.TreeParent, []byte(nil))
		}
	}
}

func (s *stateSynchGHS) receiveProposedMWOE(v lib.WeightedGraphNode) *edge {
	foundBetterProposedEdge := false
	s.foreachChildEdge(func(i int) {
		message := receiveMessageSynchGHS(v, i)
		if message != nil && (s.ProposedEdge == nil || message.MWOE.isLess(s.ProposedEdge)) {
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

func (s *stateSynchGHS) choseMWOE(v lib.WeightedGraphNode) messageType {
	if s.TreeRoot == v.GetIndex() {
		if s.ProposedEdge != nil {
			s.SelectedEdge = s.ProposedEdge.isConnected(v.GetIndex())
			return msgChooseEdge
		} else {
			return msgCompleated
		}
	}
	return nilMessage
}

func (s *stateSynchGHS) propagateDownChosenMWOE(v lib.WeightedGraphNode, message messageType) {
	switch message {
	case msgChooseEdge:
		s.foreachChildEdge(func(i int) {
			if i == s.ProposedEdgeIndex {
				sendMessageSynchGHS(v, s.ProposedEdgeIndex, &messageSynchGHS{Type: msgChooseEdge})
			} else {
				v.SendMessage(i, nil)
			}
		})
	case msgCompleated:
		s.foreachChildEdge(func(i int) {
			sendMessageSynchGHS(v, i, &messageSynchGHS{Type: msgCompleated})
		})
	default:
		s.foreachChildEdge(func(i int) {
			v.SendMessage(i, nil)
		})
	}
}

func (s *stateSynchGHS) receiveChosenMWOE(v lib.WeightedGraphNode) messageType {
	if s.TreeParent != nilParent {
		message := receiveMessageSynchGHS(v, s.TreeParent)
		if message != nil && message.Type == msgChooseEdge {
			s.SelectedEdge = s.ProposedEdge.isConnected(v.GetIndex())
		}
		if message != nil {
			return message.Type
		}
	}
	return nilMessage
}

func (s *stateSynchGHS) sendConnectComponents(v lib.WeightedGraphNode) {
	s.foreachOutgoingEdge(func(i int) {
		if s.SelectedEdge && i == s.ProposedEdgeIndex {
			s.Edges[i] = treeEdge
			sendMessageSynchGHS(v, i, &messageSynchGHS{Type: msgConnectEdge})
			log.Printf("Adding edge (%d, %d) with weight %d to MST\n", s.ProposedEdge.LargerV, s.ProposedEdge.SmallerV, s.ProposedEdge.Weight)
		} else {
			v.SendMessage(i, []byte(nil))
		}
	})
}

func (s *stateSynchGHS) electNewRoot(v lib.WeightedGraphNode) *int {
	var newRoot *int = nil
	if s.SelectedEdge {
		msg := receiveMessageSynchGHS(v, s.ProposedEdgeIndex)
		if msg != nil && msg.Type == msgConnectEdge && v.GetIndex() == s.ProposedEdge.LargerV {
			s.TreeParent = nilParent
			s.TreeRoot = v.GetIndex()
			newRoot = &s.TreeRoot
		}
	}
	s.foreachOutgoingEdge(func(i int) {
		msg := receiveMessageSynchGHS(v, i)
		if msg != nil && msg.Type == msgConnectEdge {
			s.Edges[i] = treeEdge
		}
	})
	return newRoot
}

func (s *stateSynchGHS) propagateDownNewRoot(v lib.WeightedGraphNode, newRoot *int) {
	s.foreachTreeEdge(func(i int) {
		if newRoot != nil && i != s.TreeParent {
			sendMessageSynchGHS(v, i, &messageSynchGHS{Type: msgElectNewRoot, Root: *newRoot})
		} else {
			v.SendMessage(i, nil)
		}
	})
}

func (s *stateSynchGHS) receiveNewRoot(v lib.WeightedGraphNode) *int {
	var newRoot *int = nil
	s.foreachTreeEdge(func(i int) {
		message := receiveMessageSynchGHS(v, i)
		if message != nil {
			s.TreeRoot = message.Root
			s.TreeParent = i
			newRoot = &message.Root
		}
	})
	return newRoot
}

// Initialize state and send verification messages for outgoing edges
func initializeSynchGHS(v lib.WeightedGraphNode) bool {
	s := &stateSynchGHS{
		TreeRoot:          v.GetIndex(),
		Edges:             make([]edgeStatus, 0),
		TreeParent:        nilParent,
		ProposedEdge:      nil,
		ProposedEdgeIndex: -1,
		SelectedEdge:      false,
	}
	for range v.GetOutNeighbors() {
		s.Edges = append(s.Edges, outgoingEdge)
	}
	s.sendVerifyEdges(v)
	s.Status = stateVerifyOutgoingEdges
	s.saveState(v)
	return false
}

// The algorithm requires each node to now the number of nodes in network
func processSynchGHS(v lib.WeightedGraphNode, n int) bool {
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
			finished = msg == msgCompleated
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
		finished = msg == msgCompleated

	// Electing new root. For every new tree there is a single edge e which was a MWOE for
	// two tree in the previous round. The vertice with larger index connected to e becomes new root.
	case stateElectNewRoot:
		root := s.electNewRoot(v)
		s.propagateDownNewRoot(v, root)
		s.Status = stateBroadcaseNewRoot
		s.BroadcastRound = 0

	// Broadcasting information about new root. After n rounds all vertices in tree have received
	// information about new root. Then verify every outgoing edge if it points to different tree.
	case stateBroadcaseNewRoot:
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

func runSynchGHS(v lib.WeightedGraphNode, n int) {
	v.StartProcessing()
	finish := initializeSynchGHS(v)
	v.FinishProcessing(false)

	for !finish {
		v.StartProcessing()
		finish = processSynchGHS(v, n)
		v.FinishProcessing(finish)
	}
}

func RunSynchGHS(vertices []lib.WeightedGraphNode, synchronizer lib.Synchronizer) bool {
	for _, v := range vertices {
		go runSynchGHS(v, len(vertices))
	}
	synchronizer.Synchronize(0)
	synchronizer.GetStats()
	return verifySynchSHS(vertices)
}

func RunSynchGHSRandom(n, m, maxWeight int) bool {
	vertices, synchronizer := lib.BuildSynchronizedRandomConnectedWeightedGraph(n, m, maxWeight, lib.GetGenerator())
	return RunSynchGHS(vertices, synchronizer)
}
