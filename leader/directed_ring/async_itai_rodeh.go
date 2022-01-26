package directed_ring

import (
	"encoding/json"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
	"math/rand"
	"sync"
	"time"	
)

type APL byte
const (
	ACTIVE APL = iota
	PASSIVE
	LEADER
)

type stateAsyncItaiRodeh struct {
	Id    int
	State APL
	Round int
	MsgsSent      int  // counts all sent messages to have ClosingMsgId of the last one
	MsgsReceived  int  // counts all received messages to receive all sent from previous vertex including ClosingMsg
	ClosingMode   bool // turn off sending messages, 
	ClosingMsgId  int  // id of the last sent Msg by the previous node; there can be only one goroutine in processAsyncItaiRodeh so all message Ids are distint
}

type messageAsyncItaiRodeh struct {
	Id    int
	Round int
	Hop   int
	Bit   bool
	MsgId int
}

func setStateAsyncItaiRodeh(v lib.Node, state stateAsyncItaiRodeh) {
	data, _ := json.Marshal(state)
	v.SetState(data)
}

func getStateAsyncItaiRodeh(node lib.Node) stateAsyncItaiRodeh {
	var state stateAsyncItaiRodeh
	json.Unmarshal(node.GetState(), &state)
	return state
}

func SendMessageAndWait(v lib.Node, index int, message []byte, wg *sync.WaitGroup) { // OPTION 2
	defer wg.Done()
	v.SendMessage(index, message)
}

func sendAsyncItaiRodeh(v lib.Node, id int, round int, hop int, bit bool, msgId int, closing bool) {
	outMessage, _ := json.Marshal(messageAsyncItaiRodeh {		
		Id:    id,
		Round: round,
		Hop:   hop,
		Bit:   bit,
		MsgId: msgId,
	})

	if closing {
		//var wg sync.WaitGroup
		//wg.Add(1)
		//go SendMessageAndWait(v, 0, outMessage, &wg)
		//wg.Wait()

		// OPTION 1
		v.SendMessage(0, outMessage) // HERE: i suspect that go v.SendMessage() sends message after v.FinishProcessing(true) and it causes problems
	} else {
		go v.SendMessage(0, outMessage)
	}
}

func receiveAsyncItaiRodeh(v lib.Node) messageAsyncItaiRodeh {
	var msg messageAsyncItaiRodeh
	inMessage := v.ReceiveMessage(0)
	json.Unmarshal(inMessage, &msg)
	return msg
}

func initializeAsyncItaiRodeh(v lib.Node) {
	setStateAsyncItaiRodeh(v, stateAsyncItaiRodeh {
		Id:    rand.Int()%2, // TODO: remove %2
		State: ACTIVE,
		Round: 1,
		MsgsSent:     1, // TODO: make here 0 after adding increment to sending function
		MsgsReceived: 0,
		ClosingMode:  false,
		ClosingMsgId: 0, 
	})	
	state := getStateAsyncItaiRodeh(v)	
	sendAsyncItaiRodeh(v, state.Id, state.Round, 1, true, state.MsgsSent, false)
}

func processAsyncItaiRodeh(v lib.Node) bool {
	msg := receiveAsyncItaiRodeh(v)
	state := getStateAsyncItaiRodeh(v)
	state.MsgsReceived++
	setStateAsyncItaiRodeh(v, state)

	if msg.Round == -1 { // closing msg received (leader elected)
		state.ClosingMode = true
		state.ClosingMsgId = msg.MsgId				
		log.Println("                round=-1 in ", v.GetIndex(), " , ", state.MsgsReceived, " , ", state.ClosingMsgId) // TODO: remove this
		
		if state.State != LEADER {
			state.MsgsSent++
			setStateAsyncItaiRodeh(v, state)
			sendAsyncItaiRodeh(v, msg.Id, -1, msg.Hop + 1, false, state.MsgsSent, true) // closing round has the last id sent from this vertex and nothing else will be sent
		}
	}
	
	if state.ClosingMode {
		if state.MsgsReceived == state.ClosingMsgId { // waiting for all incoming messages if closing message is not the last one (no FIFO)
			// OPOTION 3
			//time.Sleep(100 * time.Millisecond) // HERE: i suspect that go v.SendMessage() sends message after v.FinishProcessing(true) and it causes problems
			return false
		}
	} else {
		if state.State == PASSIVE {
			state.MsgsSent++ // TODO: move this and next line to send function cuz repetitions
			setStateAsyncItaiRodeh(v, state)
			sendAsyncItaiRodeh(v, msg.Id, msg.Round, msg.Hop + 1, msg.Bit, state.MsgsSent, false)
		} else if state.State == ACTIVE {
			if msg.Hop == v.GetSize() && msg.Bit {
				state.State = LEADER
				state.MsgsSent++
				state.ClosingMode = true
				setStateAsyncItaiRodeh(v, state)
				sendAsyncItaiRodeh(v, msg.Id, -1, 1, false, state.MsgsSent, true) // closing round has the last id sent from this vertex and nothing else will be sent
				log.Println("                                Elected node", v.GetIndex(), "as a leader")
			} else if msg.Hop == v.GetSize() && !msg.Bit {
				state.Id = rand.Int()%2 // TODO: remove %2
				state.Round++
				state.MsgsSent++
				setStateAsyncItaiRodeh(v, state)
				sendAsyncItaiRodeh(v, state.Id, state.Round, 1, true, state.MsgsSent, false)
			} else if msg.Round == state.Round && msg.Id == state.Id && msg.Hop < v.GetSize() { // msg.Hop < v.GetSize() is always true, so why is it in paper?
				state.MsgsSent++
				setStateAsyncItaiRodeh(v, state)
				sendAsyncItaiRodeh(v, msg.Id, msg.Round, msg.Hop + 1, false, state.MsgsSent, false)
			} else if msg.Round > state.Round || msg.Round == state.Round && msg.Id > state.Id {
				state.MsgsSent++
				setStateAsyncItaiRodeh(v, state)
				sendAsyncItaiRodeh(v, msg.Id, msg.Round, msg.Hop + 1, msg.Bit, state.MsgsSent, false)
			} // else purge msg
		}	
	}	
	return true
}

func runAsyncItaiRodeh(v lib.Node) {
	v.StartProcessing()
	initializeAsyncItaiRodeh(v)
	for ; processAsyncItaiRodeh(v); {}
	v.FinishProcessing(true)
}

func RunAsyncItaiRodeh(vertices []lib.Node, runner lib.Runner) (int, int) {
    rand.Seed(time.Now().UnixNano())

	for _, v := range vertices {
		log.Println("Running node", v.GetIndex())
		go runAsyncItaiRodeh(v)
	}

	runner.Run()
	//checkSingleLeaderElected(vertices) // TODO

	return runner.GetStats()
}
