package async_korach_moran_zaks

import (
	"encoding/json"
	"log"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

type state struct {
	Role        int
	Phase       int
	Rivals      map[int]int
	Vassals     map[int]int
	Overlord    int
	King        int
	AskedBefore message
	Queue       []message
}

type message struct {
	Flag  int
	Phase int
	King  int
	Index int
}

const (
	ASK = iota
	ACCEPT
	UPDATE
	YOUR_CITIZEN
	I_AM_THE_KING
)

const (
	KING = iota
	CITIZEN_MAIN
	CITIZEN_UPDATE
	CITIZEN_ASK_UPDATE
)

/************************************************************/
/*                           KING                           */
/************************************************************/

func processKing(node lib.Node, received_index int, received_message message) bool {
	send := func(index int, message []byte) {
		node.SendMessage(index, message)
	}

	state := getState(node)

	defer setState(node, &state)

	message_type := received_message.Flag

	switch message_type {
	case ASK:
		if state.Phase > received_message.Phase || (state.Phase == received_message.Phase && state.King > received_message.King) {
			return false
		} else {
			_, ok := state.Vassals[received_index]
			if ok {
				delete(state.Vassals, received_index)
			}
			state.Overlord = received_index
			state.Role = CITIZEN_UPDATE
			send(received_index, createMessage(ACCEPT, state.Phase, state.King))
			return true
		}
	case YOUR_CITIZEN:
	case ACCEPT:
		state.Vassals[received_index] = received_index
		if state.Phase > received_message.Phase {
			send(received_index, createMessage(UPDATE, state.Phase, state.King))
		} else {
			state.Phase++
			for v := range state.Vassals {
				send(v, createMessage(UPDATE, state.Phase, state.King))
			}
		}
	}

	//len(state.Rivals) == 0 means that every ASK provoked the response -> we are the King of Kings
	if state.Role != KING || len(state.Rivals) == 0 {
		return true
	}
	index := getSomeKey(state.Rivals)
	delete(state.Rivals, index)
	send(index, createMessage(ASK, state.Phase, node.GetIndex()))
	return false
}

func iAmTheKing(node lib.Node) {
	king_message := message{I_AM_THE_KING, I_AM_THE_KING, node.GetIndex(), 0}
	the_king, _ := json.Marshal(king_message)
	for v := range getState(node).Vassals {
		node.SendMessage(v, the_king)
	}
}

/************************************************************/
/*                         CITIZEN                          */
/************************************************************/

// Entering that function means that we were dethroned by some more powerful King - we have saved that ASK message in Queue
func processCitizen(node lib.Node, received_index int, received_message message) bool {
	send := func(index int, message []byte) {
		node.SendMessage(index, message)
	}

	message_type := received_message.Flag
	state := getState(node)
	defer setState(node, &state)

	switch message_type {
	case ASK:
		setState(node, &state)
		processAsk(node, received_index, received_message)
		state = getState(node)
	case ACCEPT:
		setState(node, &state)
		processAcceptOld(node, received_index, received_message)
		state = getState(node)
	case UPDATE:
		setState(node, &state)
		processUpdate(node, received_message)
		state = getState(node)
	case I_AM_THE_KING:
		for v := range state.Vassals {
			send(v, createMessage(I_AM_THE_KING, received_message.Phase, received_message.King))
		}
		return false
	}
	return true
}

// The node is waiting for UPDATE message from his king
func citizenUpdate(node lib.Node, received_index int, received_message message) bool {
	state := getState(node)
	defer setState(node, &state)

	message_type := received_message.Flag
	if message_type == UPDATE {
		setState(node, &state)
		processUpdate(node, received_message)
		state = getState(node)
		return true
	} else {
		received_message.Index = received_index
		state.Queue = append(state.Queue, received_message)
		return false
	}
}

// The node is waiting for UPDATE message to detemine if the sender of previous ASK message was his king
// Or to change his King to the new one
func citizenAskUpdate(node lib.Node, received_index int, received_message message) bool {
	send := func(index int, message []byte) {
		node.SendMessage(index, message)
	}

	state := getState(node)

	defer setState(node, &state)

	//we need to remember who send ASK to us so we can respond after we discuss the matter with our king
	ask_index := state.AskedBefore.Index
	ask_message := state.AskedBefore

	message_type := received_message.Flag

	if message_type == UPDATE {

		setState(node, &state)
		processUpdate(node, received_message)
		state = getState(node)

		if (state.Phase == ask_message.Phase && state.King == ask_message.King) && ask_index != state.Overlord {
			//we were the vassal of a king who became vassal of another King
			//that King send us the ASK message
			//the UPDATE message came after the ASK message
			send(ask_index, createMessage(YOUR_CITIZEN, state.Phase, state.King))
			state.Role = CITIZEN_MAIN
			return true
		}
	} else if message_type == ACCEPT {
		setState(node, &state)
		if received_index == state.Overlord {
			processAcceptNew(node, received_index, received_message)
			state = getState(node)
			return true
		} else {
			received_message.Index = received_index
			state.Queue = append(state.Queue, received_message)
		}
	} else {
		received_message.Index = received_index
		state.Queue = append(state.Queue, received_message)
	}

	if received_message.Phase > ask_message.Phase || (received_message.Phase == ask_message.Phase && received_message.King > ask_message.King) {
		state.Role = CITIZEN_MAIN
		return true
	}

	return false
}

func processAsk(node lib.Node, ask_index int, ask_message message) {

	send := func(index int, message []byte) {
		node.SendMessage(index, message)
	}

	state := getState(node)

	defer setState(node, &state)

	//we need to remember who send ASK to us so we can respond after we discuss the matter with our king
	ask_message.Index = ask_index
	state.AskedBefore = ask_message

	if state.Phase < ask_message.Phase || (state.Phase == ask_message.Phase && state.King < ask_message.King) {
		send(state.Overlord, createMessage(ask_message.Flag, ask_message.Phase, ask_message.King))
		state.Role = CITIZEN_ASK_UPDATE
	} else if state.Phase == ask_message.Phase && state.King == ask_message.King {
		if ask_index != state.Overlord {
			//we were the vassal of a king who became vassal of another King
			//that King send us the ASK message
			//the UPDATE message came before the ASK message
			send(ask_index, createMessage(YOUR_CITIZEN, state.Phase, state.King))
		}
	}
}

// Some citizen accepted to become our vassal - response to our ASK from good old times when we were still a king
func processAcceptOld(node lib.Node, received_index int, received_message message) {
	state := getState(node)
	state.Vassals[received_index] = received_index
	node.SendMessage(received_index, createMessage(UPDATE, state.Phase, state.King))
	setState(node, &state)
}

// Our old King is now a citizen so our overlord is now our vassal
func processAcceptNew(node lib.Node, received_index int, received_message message) {
	state := getState(node)

	state.Vassals[received_index] = received_index
	//the new overlord might have been our vassal before
	_, ok := state.Vassals[state.AskedBefore.Index]
	if ok {
		delete(state.Vassals, state.AskedBefore.Index)
	}
	state.Overlord = state.AskedBefore.Index

	defer setState(node, &state)

	node.SendMessage(state.Overlord, createMessage(ACCEPT, state.Phase, state.King))

	state.Role = CITIZEN_UPDATE
}

// UPDATE message is only sent when something changes
func processUpdate(node lib.Node, received_message message) {
	state := getState(node)
	state.Phase = received_message.Phase
	state.King = received_message.King

	for v := range state.Vassals {
		node.SendMessage(v, createMessage(UPDATE, state.Phase, state.King))
	}
	setState(node, &state)
}

/************************************************************/
/*                         	 MAIN                           */
/************************************************************/

func process(node lib.Node) bool {
	state := getState(node)

	defer setState(node, &state)
	var received_index int
	var received_message message

	//receive the message or take it from the queue, This is the only place where nodes receives messages
	if state.Role == CITIZEN_UPDATE || state.Role == CITIZEN_ASK_UPDATE || len(state.Queue) == 0 {
		received_index, received_message = receiveMessage(node)
	} else {
		received_message = state.Queue[0]
		received_index = state.Queue[0].Index
		state.Queue = state.Queue[1:]
	}

	setState(node, &state)

	switch state.Role {
	case KING:
		if processKing(node, received_index, received_message) {
			state = getState(node)
			if state.Role == KING {
				iAmTheKing(node)
				return false
			} else {
				return true
			}
		} else {
			state = getState(node)
			return true
		}
	case CITIZEN_MAIN:
		if processCitizen(node, received_index, received_message) {
			state = getState(node)
			return true
		} else {
			state = getState(node)
			return false
		}
	case CITIZEN_UPDATE:
		if citizenUpdate(node, received_index, received_message) {
			state = getState(node)
			state.Role = CITIZEN_MAIN
		} else {
			state = getState(node)
		}
	case CITIZEN_ASK_UPDATE:
		if citizenAskUpdate(node, received_index, received_message) {
			state = getState(node)
		} else {
			state = getState(node)
		}
	}

	return true
}

func initialize(node lib.Node) {
	n := node.GetSize() - 1
	new_state := state{}
	new_state.Rivals = make(map[int]int)
	new_state.Vassals = make(map[int]int)
	for i := 0; i < n; i++ {
		new_state.Rivals[i] = i
	}
	new_state.Phase = 0
	new_state.Role = KING
	new_state.Overlord = node.GetIndex()
	new_state.King = node.GetIndex()
	new_state.Queue = make([]message, 0)

	//send first message to begin algorithm
	index := getSomeKey(new_state.Rivals)
	delete(new_state.Rivals, index)
	node.SendMessage(index, createMessage(ASK, new_state.Phase, node.GetIndex()))

	setState(node, &new_state)
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
	checkConqueror(vertices)
	return runner.GetStats()
}

/************************************************************/
/*                      HELPER FUNCTIONS                    */
/************************************************************/

func receiveMessage(node lib.Node) (int, message) {
	received_index, received_bytes := node.ReceiveAnyMessage()
	received_message := message{}
	json.Unmarshal(received_bytes, &received_message)
	return received_index, received_message
}

func getSomeKey(m map[int]int) int {
	for k := range m {
		return k
	}
	return -1
}

func createMessage(flag int, phase int, king int) []byte {
	send_message, _ := json.Marshal(message{flag, phase, king, 0})
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

func checkConqueror(nodes []lib.Node) {
	king := getState(nodes[0]).King
	for _, node := range nodes {
		if getState(node).King != king {
			panic("Multiple conquerors spotted")
		}
	}
	log.Println("The King of Kings is ", king)
}
