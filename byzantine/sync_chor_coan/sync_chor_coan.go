package sync_chor_coan

import (
	"encoding/json"
	"log"
	"math"
	"math/rand"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

const (
	ZERO     = 0
	ONE      = 1
	ANY_NONE = iota
)

type state struct {
	AgreedValue    int
	CurrentValue   int
	Toss           int
	T              int //max number of byzantine processors
	N              int //number of all processors
	G              int //size of each (but the last) group
	NumberOfGroups int
	Byzantine      bool
}

type Message struct {
	Current int
	Toss    int
}

func Run(nodes []lib.Node, synchronizer lib.Synchronizer, T int, inputs []int, faultyBehavior func(lib.Node, int)) (bool, int) {
	if faultyBehavior == nil {
		faultyBehavior = runByzantine
	}
	for i, node := range nodes {
		if inputs[i] == -1 {
			log.Println("Running node", node.GetIndex(), "as byzantine")
			go faultyBehavior(node, T)
		} else {
			log.Println("Running node", node.GetIndex())
			go run(node, T, inputs[i])
		}
	}

	synchronizer.Synchronize(0)
	messages, time := synchronizer.GetStats()
	log.Println("messages:", messages)
	log.Println("time:", time)
	return checkConsensus(nodes)
}

func tossCoin() int {
	return rand.Intn(2)
}

func checkConsensus(nodes []lib.Node) (bool, int) {
	decided := [...]bool{false, false}
	for _, node := range nodes {
		state := getState(node)
		if !state.Byzantine {
			if state.AgreedValue == ANY_NONE {
				log.Println("At least one non-byzantine node did not agreed on any value")
				return false, -1
			} else {
				decided[state.AgreedValue] = true
			}
		}
	}

	if decided[0] && decided[1] {
		log.Println("Non-byzantine nodes agreed on different values")
		return false, -1
	} else if decided[0] {
		log.Println("All non-byzantine nodes agreed on value 0")
		return true, 0
	} else if decided[1] {
		log.Println("All non-byzantine nodes agreed on value 1")
		return true, 1
	} else {
		log.Println("Non-byzantine nodes did not agree on any value")
		return false, -1
	}
}

func handleRoundOne(node lib.Node, messages []*Message) {
	zeros, ones := 0, 0

	for _, msg := range messages {
		if msg.Current == ZERO {
			zeros += 1
		} else if msg.Current == ONE {
			ones += 1
		}
	}

	state := getState(node)

	if zeros >= state.N-state.T {
		state.CurrentValue = ZERO
	} else if ones >= state.N-state.T {
		state.CurrentValue = ONE
	} else {
		state.CurrentValue = ANY_NONE
	}

	setState(node, &state)
}

func handleCoinToss(node lib.Node, epoch int) {
	state := getState(node)
	if epoch%state.G == 0 {
		state.Toss = tossCoin()
	} else {
		state.Toss = ANY_NONE
	}
}

func handleRoundTwo(node lib.Node, messages []*Message, epoch int) bool {
	state := getState(node)

	zerosAns, onesAns := 0, 0
	zerosToss, onesToss := 0, 0

	for id, msg := range messages {
		if msg.Current == ZERO {
			zerosAns += 1
		} else if msg.Current == ONE {
			onesAns += 1
		}

		if id/state.G == (epoch % state.NumberOfGroups) { //checking whether sender was really in a coin throwing group
			if msg.Toss == ZERO {
				zerosToss += 1
			} else if msg.Toss == ONE {
				onesToss += 1
			}
		}
	}

	var ans, num, ansToss int

	if zerosAns > onesAns {
		num, ans = zerosAns, ZERO
	} else {
		num, ans = onesAns, ONE
	}

	if zerosToss > onesToss {
		ansToss = ZERO
	} else {
		ansToss = ONE
	}

	if num >= state.N-state.T {
		state.AgreedValue = ans
		setState(node, &state)
		return true
	} else if num >= state.T+1 {
		state.CurrentValue = ans
	} else {
		state.CurrentValue = ansToss
	}
	setState(node, &state)
	return false
}

//------------------------------Regular processor------------------------------------

func initialize(node lib.Node, T int, input int) {
	g := int(math.Log2(float64(node.GetSize())))
	setState(node, &state{
		AgreedValue:    ANY_NONE,
		CurrentValue:   input,
		Toss:           ANY_NONE,
		Byzantine:      false,
		T:              T,
		G:              g,
		N:              node.GetSize(),
		NumberOfGroups: node.GetSize() / g,
	})
}

func run(node lib.Node, T int, input int) {
	node.StartProcessing()
	initialize(node, T, input)

	epoch := 0
	for {
		broadcast(node, &Message{getState(node).CurrentValue, ANY_NONE})
		messages := receive(node)
		handleRoundOne(node, messages)
		handleCoinToss(node, epoch)
		broadcast(node, &Message{getState(node).CurrentValue, getState(node).Toss})
		messages = receive(node)
		if handleRoundTwo(node, messages, epoch) {
			break
		}
		epoch += 1
	}
	node.FinishProcessing(true)
}

//------------------------------Byzantine processor------------------------------------

func initializeByzantine(node lib.Node, T int) {
	initialize(node, T, tossCoin())
	state := getState(node)
	state.Byzantine = true
	setState(node, &state)
}

func runByzantine(node lib.Node, T int) { //random byzantine
	node.StartProcessing()
	initializeByzantine(node, T)

	epoch := 0
	for {
		broadcast(node, &Message{tossCoin(), ANY_NONE}) //send random value
		receive(node)
		//handleRoundOne(node, messages) simply ignore round one
		handleCoinToss(node, epoch)
		broadcast(node, &Message{tossCoin(), getState(node).Toss}) //send random value
		messages := receive(node)
		if handleRoundTwo(node, messages, epoch) {
			break
		}
		epoch += 1
	}
	node.FinishProcessing(true)
}

//------------------------------Helper functions------------------------------------

func getState(node lib.Node) state {
	var state state
	json.Unmarshal(node.GetState(), &state)
	return state
}

func setState(node lib.Node, state *state) {
	encodedState, _ := json.Marshal(*state)
	node.SetState(encodedState)
}

func send(node lib.Node, msg *Message, dest int) {
	msgAsJson, _ := json.Marshal(msg)
	node.SendMessage(dest, msgAsJson)
}

func broadcast(node lib.Node, msg *Message) {
	for i := 0; i < node.GetOutChannelsCount(); i++ {
		send(node, msg, i)
	}
}

func receive(node lib.Node) []*Message {
	msgs := make([]*Message, node.GetInChannelsCount())

	for i := 0; i < node.GetInChannelsCount(); i++ {
		msgAsJson := node.ReceiveMessage(i)
		if msgAsJson != nil {
			msgs[i] = &Message{}
			json.Unmarshal(msgAsJson, msgs[i])
		}
	}
	return msgs
}
