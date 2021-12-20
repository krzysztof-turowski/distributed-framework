package directed_hypercube

import (
	"encoding/json"
	"lib"
	"log"
)

/* HYPERELECT TYPES */

type StatusType int

const (
	duellist StatusType = iota
	defeated
	leader
	follower
)

type MessageType int

const (
	match MessageType = iota
	follow
)

type MessageHyperelect struct {
	Type        MessageType
	Value       int
	Stage       int
	Source      []bool
	Destination []int
}

type StateHypercube struct {
	Status       StatusType
	Stage        int
	Delayed      []*MessageHyperelect
	Pending      [][]*MessageHyperelect
	NextDuellist []int
}

/* STATE HYPERELECT METHODS */

func newStateHyperelect(dim int) *StateHypercube {
	var st StateHypercube
	st.Status = duellist
	st.Stage = 0
	st.Delayed = make([]*MessageHyperelect, dim)
	st.Pending = make([][]*MessageHyperelect, dim)
	for i := range st.Pending {
		st.Pending[i] = make([]*MessageHyperelect, 0)
	}
	st.NextDuellist = nil
	return &st
}

func getStateHyperelect(v lib.Node) *StateHypercube {
	var st StateHypercube
	json.Unmarshal(v.GetState(), &st)
	return &st
}

func setStateHyperelect(v lib.Node, st *StateHypercube) {
	data, _ := json.Marshal(*st)
	v.SetState(data)
}

func addDelayed(st *StateHypercube, msg *MessageHyperelect) {
	st.Delayed[msg.Stage] = msg
}

func addPending(index int, st *StateHypercube, msg *MessageHyperelect) {
	st.Pending[index] = append(st.Pending[index], msg)
}

func clearPending(st *StateHypercube) {
	for i := range st.Pending {
		st.Pending[i] = nil
	}
}

/* MESSAGE HYPERELECT METHODS */

func newMessageHyperelect(value int, stage int, dim int) *MessageHyperelect {
	msg := MessageHyperelect{
		Type:        match,
		Value:       value,
		Stage:       stage,
		Source:      make([]bool, dim),
		Destination: make([]int, 0),
	}
	msg.Source[stage] = true
	return &msg
}

func sendMessageHyperelect(v lib.Node, index int, msg *MessageHyperelect) {
	if msg != nil {
		out, _ := json.Marshal(*msg)
		v.SendMessage(index, out)
	} else {
		v.SendMessage(index, nil)
	}
}

func receiveMessageHyperelect(v lib.Node, index int) *MessageHyperelect {
	var msg MessageHyperelect
	in := v.ReceiveMessage(index)

	if in != nil {
		json.Unmarshal(in, &msg)
		return &msg
	}
	return nil
}

func sendPendingMessages(v lib.Node, st *StateHypercube) {
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		if len(st.Pending[i]) > 0 {
			sendMessageHyperelect(v, i, st.Pending[i][0])
			st.Pending[i] = st.Pending[i][1:]
		} else {
			sendMessageHyperelect(v, i, nil)
		}
	}
}

func forwardMessage(st *StateHypercube, msg *MessageHyperelect) {
	if len(msg.Destination) == 0 {
		msg.Destination = st.NextDuellist
		for _, d := range st.NextDuellist {
			msg.Source[d] = !msg.Source[d]
		}
	}

	index := msg.Destination[0]
	msg.Destination = msg.Destination[1:]
	addPending(index, st, msg)
}

func clearIncomingMessages(v lib.Node, index int) {
	for ; index < v.GetInChannelsCount(); index++ {
		receiveMessageHyperelect(v, index)
	}
}

/* HYPERELECT METHODS */

func initializeHypercube(v lib.Node) bool {
	st := newStateHyperelect(v.GetOutChannelsCount())
	msg := newMessageHyperelect(v.GetIndex(), st.Stage, v.GetOutChannelsCount())

	addPending(0, st, msg)
	sendPendingMessages(v, st)
	setStateHyperelect(v, st)

	return false
}

func processFollow(v lib.Node, index int, st *StateHypercube) bool {
	clearIncomingMessages(v, index)
	clearPending(st)
	if st.Stage > 0 {
		addPending(st.Stage-1, st, &MessageHyperelect{Type: follow})
	}

	return st.Stage == 0
}

func processDefeated(v lib.Node, index int, st *StateHypercube) bool {
	for i := index; i < v.GetInChannelsCount(); i++ {
		msg := receiveMessageHyperelect(v, i)
		if msg != nil {
			if msg.Type == match {
				forwardMessage(st, msg)
			} else {
				st.Status = follower
				st.Stage = i
				return processFollow(v, i+1, st)
			}
		}
	}

	return false
}

func processDuelWin(v lib.Node, index int, st *StateHypercube) bool {
	st.Stage++

	if st.Stage == v.GetOutChannelsCount() {
		st.Status = leader
		clearIncomingMessages(v, index)
		clearPending(st)
		addPending(v.GetOutChannelsCount()-1, st, &MessageHyperelect{Type: follow})
		return true
	}

	m := newMessageHyperelect(v.GetIndex(), st.Stage, v.GetOutChannelsCount())
	addPending(st.Stage, st, m)

	return false
}

func processDuelLose(v lib.Node, index int, st *StateHypercube, msg *MessageHyperelect) bool {
	st.Status = defeated
	for i := range msg.Source {
		if msg.Source[i] {
			st.NextDuellist = append(st.NextDuellist, i)
		}
	}
	for i := st.Stage; i < len(st.Delayed); i++ {
		if st.Delayed[i] != nil {
			forwardMessage(st, st.Delayed[i])
			st.Delayed[i] = nil
		}
	}
	return processDefeated(v, index, st)
}

func hyperelectDuellist(v lib.Node, st *StateHypercube) bool {
	msg := st.Delayed[st.Stage]
	st.Delayed[st.Stage] = nil
	if msg != nil {
		if msg.Value > v.GetIndex() {
			if processDuelWin(v, 0, st) {
				return false
			}
		} else {
			return processDuelLose(v, 0, st, msg)
		}
	}

	for i := 0; i < v.GetInChannelsCount(); i++ {
		msg = receiveMessageHyperelect(v, i)

		if msg != nil {
			if msg.Type == match {
				if msg.Stage > st.Stage {
					addDelayed(st, msg)
				} else if msg.Value > v.GetIndex() {
					if processDuelWin(v, i+1, st) {
						return false
					}
				} else {
					return processDuelLose(v, i+1, st, msg)
				}
			} else {
				st.Status = follower
				st.Stage = i
				return processFollow(v, i+1, st)
			}
		}
	}

	return false
}

func hyperelectDefeated(v lib.Node, st *StateHypercube) bool {
	return processDefeated(v, 0, st)
}

func hyperelectClose(v lib.Node, st *StateHypercube) bool {
	st.Stage--
	return processFollow(v, 0, st)
}

func hyperelect(v lib.Node) bool {
	st := getStateHyperelect(v)
	finish := false

	if st.Status == duellist {
		finish = hyperelectDuellist(v, st)
	} else if st.Status == defeated {
		finish = hyperelectDefeated(v, st)
	} else {
		finish = hyperelectClose(v, st)
	}

	sendPendingMessages(v, st)
	setStateHyperelect(v, st)

	return finish
}

func runHyperelect(v lib.Node) {
	v.StartProcessing()
	finish := initializeHypercube(v)
	v.FinishProcessing(finish)

	for !finish {
		v.StartProcessing()
		finish = hyperelect(v)
		v.FinishProcessing(finish)
	}
}

func checkHyperelect(vertices []lib.Node) {
	var lead lib.Node
	min := vertices[0].GetIndex()
	for _, v := range vertices {
		st := getStateHyperelect(v)
		if min > v.GetIndex() {
			min = v.GetIndex()
		}

		if st.Status == leader {
			if lead != nil {
				panic("Found multiple leaders")
			} else {
				lead = v
			}
		} else if st.Status != follower {
			panic("Not all nodes finished")
		}
	}

	if min != lead.GetIndex() {
		panic("Found leader is not the expected one")
	}
}

func RunHyperelect(vertices []lib.Node, synchronizer lib.Synchronizer) {
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runHyperelect(v)
	}
	synchronizer.Synchronize(0)
	synchronizer.GetStats()
	checkHyperelect(vertices)
}
