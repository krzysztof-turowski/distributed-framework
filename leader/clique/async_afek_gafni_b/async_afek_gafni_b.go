package async_afek_gafni_b

import (
	"encoding/json"
	"log"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

type state struct {
	Leader           int
	Untraversed      map[int]int
	Father           int
	Status           int
	Owner_ID         int
	Level            int
	Potential_Father int
	Ending           bool
	Counter          int
	Queue            []message
}

type message struct {
	Index  int
	Level  int
	ID     int
	Type   int
	Sender int
}

const (
	CANDIDATE = iota
	ORDINARY
	WAIT_FOR_ANSWER
)

const (
	ARRIVE = iota
	ANSWER_ACCEPT
	ANSWER_ACCEPT_CND
	ANSWER_DENY_CND
	ASK
	DEAD
	LEADER
	END
)

const (
	NOT_SET   = -1
	BEGINNING = 0
)

/************************************************************/
/*                       CANDIDATE                          */
/************************************************************/

// begin ending sequence
func announceLeader(node lib.Node) {
	n := node.GetSize() - 1
	for i := 0; i < n; i++ {
		node.SendMessage(i, createMessage(0, node.GetIndex(), LEADER, node.GetIndex()))
	}
}

// every node is ready to end -> there won't be another source of sending messages
func announceEnd(node lib.Node) {
	n := node.GetSize() - 1
	for i := 0; i < n; i++ {
		node.SendMessage(i, createMessage(0, node.GetIndex(), END, node.GetIndex()))
	}
}

func candidate(node lib.Node, received_index int, received_message message) bool {

	send := func(index int, message []byte) {
		node.SendMessage(index, message)
	}

	state := getState(node)
	defer setState(node, &state)
	message_type := received_message.Type

	if received_message.ID == node.GetIndex() && state.Status == CANDIDATE {
		state.Level++
		delete(state.Untraversed, received_index)
	} else {
		if state.Status == ORDINARY {
			if message_type == ASK { //some node still believes that this node is candidate so we need to update that information
				send(received_index, createMessage(received_message.Level, received_message.ID, DEAD, node.GetIndex()))
				return true
			}
		}
		if message_type == ASK {
			if state.Level < received_message.Level || (state.Level == received_message.Level && node.GetIndex() < received_message.ID) {
				send(received_index, createMessage(received_message.Level, received_message.ID, ANSWER_ACCEPT_CND, node.GetIndex()))
				state.Status = ORDINARY
			} else { //to avoid deadlock
				send(received_index, createMessage(received_message.Level, received_message.ID, ANSWER_DENY_CND, node.GetIndex()))
			}
		}
		return true
	}

	if len(state.Untraversed) == 0 {
		return false
	}
	index := getSomeKey(state.Untraversed)
	delete(state.Untraversed, index)
	send(index, createMessage(state.Level, node.GetIndex(), ARRIVE, node.GetIndex()))
	return true
}

/************************************************************/
/*                       ORDINARY                           */
/************************************************************/

func ordinary(node lib.Node, received_index int, received_message message) {

	send := func(index int, message []byte) {
		node.SendMessage(index, message)
	}

	state := getState(node)
	defer setState(node, &state)

	if state.Level > received_message.Level || (state.Level == received_message.Level && state.Owner_ID > received_message.ID) {
		//discard the message -> the sender is waiting to be killed because he can't progress his candidate status
		return
	} else if state.Level < received_message.Level || (state.Level == received_message.Level && state.Owner_ID < received_message.ID) {
		state.Status = ORDINARY
		state.Level = received_message.Level
		if state.Owner_ID == received_message.ID { //ARRIVE MESSAGE WAS SENT BY OUR OWNER
			state.Level = received_message.Level
			state.Status = ORDINARY
			send(received_index, createMessage(received_message.Level, received_message.ID, ANSWER_ACCEPT, node.GetIndex()))
			return
		}

		state.Potential_Father = received_index
		state.Owner_ID = received_message.ID

		if state.Father == NOT_SET {
			state.Father = state.Potential_Father
			send(state.Father, createMessage(received_message.Level, received_message.ID, ANSWER_ACCEPT, node.GetIndex()))
			return
		}

		//Since we only can send one message to channel we need to wait for response to avoid the deadlock
		state.Status = WAIT_FOR_ANSWER
		send(state.Father, createMessage(received_message.Level, received_message.ID, ASK, node.GetIndex()))
	}
}

func ordinaryWait(node lib.Node, received_index int, received_message message) {
	send := func(index int, message []byte) {
		node.SendMessage(index, message)
	}

	state := getState(node)
	defer setState(node, &state)

	message_type := received_message.Type
	if message_type == ANSWER_ACCEPT_CND || message_type == DEAD {
		//the father was killed by our ASK messages (ANSWER_ACCEPT_CND) or it was killed before
		state.Father = state.Potential_Father
		state.Status = ORDINARY
		send(state.Father, createMessage(received_message.Level, received_message.ID, ANSWER_ACCEPT, node.GetIndex()))
	} else if message_type == ANSWER_DENY_CND {
		//the father denied the ASK message
		state.Status = ORDINARY
	} else {
		//we are waiting for the response from the father - put into the queue the message
		state.Queue = append(state.Queue, message{Index: received_index, Level: received_message.Level, ID: received_message.ID, Type: received_message.Type, Sender: received_message.Sender})
	}
}

/************************************************************/
/*                         	 MAIN                           */
/************************************************************/

func process(node lib.Node) bool {
	send := func(index int, message []byte) {
		node.SendMessage(index, message)
	}

	state := getState(node)
	defer setState(node, &state)

	var received_index int
	var received_message message

	setState(node, &state)

	if len(state.Queue) == 0 || state.Status == WAIT_FOR_ANSWER {
		received_index, received_message = receiveMessage(node)
	} else {
		received_message = state.Queue[0]
		received_index = received_message.Index
		state.Queue = state.Queue[1:]
	}

	message_type := received_message.Type

	if message_type == LEADER {
		state.Leader = received_message.ID
		state.Father = received_index
		send(received_index, createMessage(received_message.Level, state.Leader, END, node.GetIndex()))
		return true
	} else if message_type == END {
		if node.GetIndex() == received_message.ID {
			state.Counter++
		} else {
			return false
		}
		if state.Counter == node.GetSize()-1 {
			announceEnd(node)
			return false
		}
	} else if state.Ending {
		return true
	} else if state.Status == WAIT_FOR_ANSWER {
		setState(node, &state)
		ordinaryWait(node, received_index, received_message)
		state = getState(node)
	} else if message_type == ARRIVE {
		setState(node, &state)
		ordinary(node, received_index, received_message)
		state = getState(node)
	} else if message_type == ANSWER_ACCEPT || message_type == ASK {
		setState(node, &state)
		if !candidate(node, received_index, received_message) {
			state = getState(node)
			if state.Status == CANDIDATE {
				announceLeader(node)
				state.Leader = node.GetIndex()
				state.Counter = 0
				state.Ending = true
				return true
			}
		} else {
			state = getState(node)
		}

	}
	return true
}

func initialize(node lib.Node) {
	n := node.GetSize() - 1
	state := state{}
	state.Level = BEGINNING
	state.Owner_ID = node.GetIndex()
	state.Father = NOT_SET
	state.Leader = NOT_SET
	state.Potential_Father = NOT_SET
	state.Status = CANDIDATE
	state.Counter = 0
	state.Ending = false
	state.Queue = make([]message, 0)
	state.Untraversed = map[int]int{}
	for i := 0; i < n; i++ {
		state.Untraversed[i] = i
	}

	index := getSomeKey(state.Untraversed)
	delete(state.Untraversed, index)
	node.SendMessage(index, createMessage(state.Level, node.GetIndex(), ARRIVE, node.GetIndex()))

	setState(node, &state)
}

func run(node lib.Node) {
	node.StartProcessing()
	initialize(node)
	for process(node) {

	}
	node.FinishProcessing(true)
}

func Run(vertices []lib.Node, runner lib.Runner) (int, int) {
	for _, v := range vertices {
		log.Println("Running node", v.GetIndex())
		go run(v)
	}
	runner.Run(true)
	checkLeader(vertices)
	return runner.GetStats()
}

/************************************************************/
/*                      HELPER FUNCTIONS                    */
/************************************************************/

func receiveMessage(node lib.Node) (int, message) {
	received_index, received_bytes := node.ReceiveAnyMessage()
	received_message := message{}
	json.Unmarshal(received_bytes, &received_message)
	//fmt.Println("STATUS:", getState(node).Status, "NODE:", node.GetIndex(), "SO", getState(node).Owner_ID, "NODELEVEL:", getState(node).Level, "|LEVEL:", received_message.Level, "|ID:", received_message.ID, "|TYPE:", received_message.Type, "|SENDER", received_message.Sender)
	return received_index, received_message
}

func getSomeKey(m map[int]int) int {
	for k := range m {
		return k
	}
	return -1
}

func createMessage(level int, id int, message_type int, sender int) []byte {
	send_message, _ := json.Marshal(message{Level: level, ID: id, Type: message_type, Sender: sender})
	return send_message
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

func checkLeader(nodes []lib.Node) {
	leader := getState(nodes[0]).Leader
	for _, node := range nodes {
		if getState(node).Leader != leader {
			panic("Multiple leaders spotted")
		}
	}
	log.Println("The Leader is ", leader)
}
