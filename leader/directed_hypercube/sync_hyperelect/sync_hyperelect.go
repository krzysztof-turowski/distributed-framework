package sync_hyperelect

import (
	"encoding/json"
	"log"

	"github.com/krzysztof-turowski/distributed-framework/lib"
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

type Message struct {
	Type        MessageType
	Value       int
	Stage       int
	Source      []bool
	Destination []int
}

type StateHypercube struct {
	Status StatusType
	Stage  int
	// messages that have greater stage than current stage
	Delayed []*Message
	// messages waiting to be sent (one per neighbor per round)
	Pending [][]*Message
	// array of dimensions to reach match winner
	NextDuellist []int
}

/* STATE HYPERELECT METHODS */

func newState(dim int) *StateHypercube {
	var st StateHypercube
	st.Status = duellist
	st.Stage = 0
	st.Delayed = make([]*Message, dim)
	st.Pending = make([][]*Message, dim)
	for i := range st.Pending {
		st.Pending[i] = make([]*Message, 0)
	}
	st.NextDuellist = nil
	return &st
}

func getState(v lib.Node) *StateHypercube {
	var st StateHypercube
	json.Unmarshal(v.GetState(), &st)
	return &st
}

func setState(v lib.Node, st *StateHypercube) {
	data, _ := json.Marshal(*st)
	v.SetState(data)
}

func addDelayedMessage(st *StateHypercube, msg *Message) {
	st.Delayed[msg.Stage] = msg
}

func addPendingMessage(index int, st *StateHypercube, msg *Message) {
	st.Pending[index] = append(st.Pending[index], msg)
}

func clearPendingMessages(st *StateHypercube) {
	for i := range st.Pending {
		st.Pending[i] = nil
	}
}

/* MESSAGE HYPERELECT METHODS */

func newMessage(value int, stage int, dim int) *Message {
	msg := Message{
		Type:        match,
		Value:       value,
		Stage:       stage,
		Source:      make([]bool, dim),
		Destination: make([]int, 0),
	}
	msg.Source[stage] = true
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

func sendPendingMessages(v lib.Node, st *StateHypercube) {
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		if len(st.Pending[i]) > 0 {
			sendMessage(v, i, st.Pending[i][0])
			st.Pending[i] = st.Pending[i][1:]
		} else {
			sendMessage(v, i, nil)
		}
	}
}

func forwardMessage(st *StateHypercube, msg *Message) {
	if len(msg.Destination) == 0 {
		msg.Destination = st.NextDuellist
		for _, d := range st.NextDuellist {
			msg.Source[d] = !msg.Source[d]
		}
	}

	index := msg.Destination[0]
	msg.Destination = msg.Destination[1:]
	addPendingMessage(index, st, msg)
}

func clearIncomingMessages(v lib.Node, index int) {
	for ; index < v.GetInChannelsCount(); index++ {
		receiveMessage(v, index)
	}
}

/* HYPERELECT METHODS */

// initializes hypercube states and sends first match message
// returns true if only one node
func initializeHypercube(v lib.Node) bool {
	st := newState(v.GetOutChannelsCount())
	log.Println("Node", v.GetIndex(), "became duellist")

	if v.GetOutChannelsCount() == 0 {
		st.Status = leader
		log.Println("Node", v.GetIndex(), "became leader")

		setState(v, st)
		return true
	}

	msg := newMessage(v.GetIndex(), st.Stage, v.GetOutChannelsCount())
	addPendingMessage(0, st, msg)
	sendPendingMessages(v, st)
	setState(v, st)

	return false
}

// processes node that is/became leader/follower
// returns true if Stage gets back to zero, i.e. all dimensions were notified
func processFollow(v lib.Node, index int, st *StateHypercube) bool {
	clearIncomingMessages(v, index)
	clearPendingMessages(st)
	if st.Stage > 0 {
		addPendingMessage(st.Stage-1, st, &Message{Type: follow})
	}

	return st.Stage == 0
}

// processes node that is/became defeated
// returns true if finished (based on follow)
func processDefeated(v lib.Node, index int, st *StateHypercube) bool {
	for i := index; i < v.GetInChannelsCount(); i++ {
		msg := receiveMessage(v, i)
		if msg != nil {
			if msg.Type == match {
				forwardMessage(st, msg)
			} else {
				st.Status = follower
				st.Stage = i
				log.Println("Node", v.GetIndex(), "became follower")

				return processFollow(v, i+1, st)
			}
		}
	}

	return false
}

// processes node that won match
// returns true if node became leader (won dim matches)
func processDuelWin(v lib.Node, index int, st *StateHypercube) bool {
	st.Stage++

	if st.Stage == v.GetOutChannelsCount() {
		st.Status = leader
		log.Println("Node", v.GetIndex(), "became leader")

		clearIncomingMessages(v, index)
		clearPendingMessages(st)
		addPendingMessage(v.GetOutChannelsCount()-1, st, &Message{Type: follow})
		return true
	}

	m := newMessage(v.GetIndex(), st.Stage, v.GetOutChannelsCount())
	addPendingMessage(st.Stage, st, m)

	return false
}

// process node that lost match
// returns true if finished (based on defeated)
func processDuelLose(v lib.Node, index int, st *StateHypercube, msg *Message) bool {
	st.Status = defeated
	log.Println("Node", v.GetIndex(), "became defeated")

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

// process node that is duellist
// return true if finished (based on received messages)
func processDuellist(v lib.Node, st *StateHypercube) bool {
	for i := 0; i < v.GetInChannelsCount(); i++ {
		msg := receiveMessage(v, i)

		if msg != nil {
			if msg.Type == match {
				if msg.Stage > st.Stage {
					addDelayedMessage(st, msg)
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
				log.Println("Node", v.GetIndex(), "became follower")

				return processFollow(v, i+1, st)
			}
		}
	}

	return false
}

// process node that started round as duellist
func hyperelectDuellist(v lib.Node, st *StateHypercube) bool {
	msg := st.Delayed[st.Stage]
	st.Delayed[st.Stage] = nil
	if msg != nil {
		if msg.Value > v.GetIndex() {
			if processDuelWin(v, 0, st) {
				return false
			} else {
				return processDuellist(v, st)
			}
		} else {
			return processDuelLose(v, 0, st, msg)
		}
	}

	return processDuellist(v, st)
}

// process node that started round as defeated
func hyperelectDefeated(v lib.Node, st *StateHypercube) bool {
	return processDefeated(v, 0, st)
}

// process node that started round as leader/follower
func hyperelectClose(v lib.Node, st *StateHypercube) bool {
	st.Stage--
	return processFollow(v, 0, st)
}

func hyperelect(v lib.Node) bool {
	st := getState(v)
	finish := false

	if st.Status == duellist {
		finish = hyperelectDuellist(v, st)
	} else if st.Status == defeated {
		finish = hyperelectDefeated(v, st)
	} else {
		finish = hyperelectClose(v, st)
	}

	sendPendingMessages(v, st)
	setState(v, st)

	return finish
}

func run(v lib.Node) {
	v.StartProcessing()
	finish := initializeHypercube(v)
	v.FinishProcessing(finish)

	for !finish {
		v.StartProcessing()
		finish = hyperelect(v)
		v.FinishProcessing(finish)
	}
}

func check(vertices []lib.Node) {
	var lead lib.Node
	min := vertices[0].GetIndex()
	for _, v := range vertices {
		st := getState(v)
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

func Run(vertices []lib.Node, synchronizer lib.Synchronizer) {
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go run(v)
	}
	synchronizer.Synchronize(0)
	synchronizer.GetStats()
	check(vertices)
}
