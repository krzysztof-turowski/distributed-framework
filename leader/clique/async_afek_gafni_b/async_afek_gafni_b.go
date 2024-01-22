package async_afek_gafni_b

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

type state struct {
	Leader           int
	Untraversed      map[int]int
	Father           int
	Killed           bool
	Owner_ID         int
	Level            int
	Potential_Father int
	//Balance          int
	//Ending           bool
}

type message struct {
	Level  int
	ID     int
	Type   int
	Sender int
}

const (
	ARRIVE = iota
	ANSWER_ACCEPT
	ANSWER_ACCEPT_CND
	ASK
	DEAD
	//DENY //just to avoid using node.IgnoreFutureMessages()
	LEADER
)

const (
	NOT_SET   = -1
	BEGINNING = 0
)

/************************************************************/
/*                       CANDIDATE                          */
/************************************************************/

func announce_leader(node lib.Node) {
	n := node.GetSize() - 1
	for i := 0; i < n; i++ {
		node.SendMessage(i, createMessage(0, node.GetIndex(), LEADER, node.GetIndex()))
	}
}

func candidate(node lib.Node, received_index int, received_message message) bool {

	//every message send in candidate function is processed by ordinary function
	send := func(index int, message []byte) {
		node.SendMessage(index, message)
	}

	state := getState(node)
	defer setState(node, &state)

	message_type := received_message.Type
	if message_type == ARRIVE {
		if received_message.Level < state.Level || (received_message.Level == state.Level && received_message.ID < node.GetIndex()) {
			//send(received_index, createMessage(state.Level, node.GetIndex(), DENY, node.GetIndex()))
			return true
		} else if received_message.Level > state.Level || (received_message.Level == state.Level && received_message.ID > node.GetIndex()) {
			state.Level = received_message.Level
			state.Father = received_index
			state.Killed = true
			state.Owner_ID = received_message.ID
			send(received_index, createMessage(received_message.Level, received_message.ID, ANSWER_ACCEPT, node.GetIndex()))
			return false
		}
	} else if message_type == ANSWER_ACCEPT {
		state.Level++
		//state.Balance--
	} else if message_type == ASK {
		if received_message.Level > state.Level || (received_message.Level == state.Level && received_message.ID > node.GetIndex()) {
			state.Level = received_message.Level
			state.Owner_ID = received_message.ID
			state.Father = received_index
			state.Killed = true
			//state.Balance++
			send(received_index, createMessage(received_message.Level, received_message.ID, ANSWER_ACCEPT_CND, node.GetIndex()))
			return false
		}
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

// the message was send from candidate part of the node
func ordinary(node lib.Node, received_index int, received_message message) {

	//every message send in ordinary function is processed by candidate function
	send := func(index int, message []byte) {
		node.SendMessage(index, message)
	}

	state := getState(node)
	defer setState(node, &state)

	message_type := received_message.Type
	if message_type == ARRIVE {
		if received_message.Level > state.Level || (received_message.Level == state.Level && received_message.ID > state.Owner_ID) {
			state.Potential_Father = received_index
			state.Level = received_message.Level
			if state.Owner_ID == received_message.ID {
				state.Father = received_index
				send(state.Father, createMessage(received_message.Level, received_message.ID, ANSWER_ACCEPT, node.GetIndex()))
			} else {
				state.Owner_ID = received_message.ID
				send(state.Father, createMessage(received_message.Level, received_message.ID, ASK, node.GetIndex()))
			}
		}
	} else if message_type == ANSWER_ACCEPT_CND {
		state.Father = state.Potential_Father
		send(state.Father, createMessage(state.Level, state.Owner_ID, ANSWER_ACCEPT, node.GetIndex()))
	} else if message_type == ASK {
		send(received_index, createMessage(received_message.Level, received_message.ID, DEAD, node.GetIndex()))
	} else if message_type == DEAD {
		state.Father = state.Potential_Father
		send(state.Father, createMessage(state.Level, state.Owner_ID, ANSWER_ACCEPT, node.GetIndex()))
	}
}

/************************************************************/
/*                         	 MAIN                           */
/************************************************************/

func process(node lib.Node) bool {
	state := getState(node)
	defer setState(node, &state)

	var received_index int
	var received_message message

	setState(node, &state)
	received_index, received_message = receiveMessage(node)

	message_type := received_message.Type

	if message_type == LEADER {
		state.Leader = received_message.ID
		return false
	} else if state.Killed {
		setState(node, &state)
		ordinary(node, received_index, received_message)
		state = getState(node)
	} else if !state.Killed {
		setState(node, &state)
		if !candidate(node, received_index, received_message) {
			state = getState(node)
			if !state.Killed {
				//fmt.Println("LEADER!", node.GetIndex())
				announce_leader(node)
				state.Leader = node.GetIndex()
				return false
			}
		}
		state = getState(node)
	}
	return true
}

func initialize(node lib.Node) {
	n := node.GetSize() - 1
	state := state{}
	state.Level = BEGINNING
	state.Owner_ID = node.GetIndex()
	state.Father = NOT_SET
	state.Potential_Father = NOT_SET
	//state.Balance = 0
	state.Killed = false
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
	//fmt.Println(node.GetIndex())
	initialize(node)
	for process(node) {

	}
	//node.IgnoreFutureMessages()
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
	//fmt.Println("KILLED:", getState(node).Killed, "NODE:", node.GetIndex(), "SO", getState(node).Owner_ID, "NODELEVEL:", getState(node).Level, "|LEVEL:", received_message.Level, "|ID:", received_message.ID, "|TYPE:", received_message.Type, "|SENDER", received_message.Sender)
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
	fmt.Println("The Leader is ", leader)
}
