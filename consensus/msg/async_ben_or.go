package ben_or

import (
	"encoding/json"
	"log"
	"math/rand"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

const (
	msgDecided   = 1
	msgUndecided = 0
)

type messageTypeBenOr int

const (
	msgTypeOne messageTypeBenOr = iota
	msgTypeTwo
)

type messageBenOr struct {
	Type    messageTypeBenOr
	Round   int
	Value   byte
	Decided int
}

type stateBenOr struct {
	StartingBit byte
	EndingBit   byte
	N           int
	T           int
	// round -> [ids that have already sent this round]
	ReceivedOnes map[int]map[int]bool

	Ones0 map[int]int
	Ones1 map[int]int

	// round -> [ids that have already sent this round]
	ReceivedTwos map[int]map[int]bool

	TwosU0 map[int]int
	TwosU1 map[int]int
	TwosD0 map[int]int
	TwosD1 map[int]int
	// used to discard messages coming from previous rounds
	RecentRound int
}

// returns number of type 1 messages with value 'val'
func getTypeOne(state *stateBenOr, val byte) int {
	if val == 0 {
		return state.Ones0[state.RecentRound]
	} else {
		return state.Ones1[state.RecentRound]
	}
}

func getAllTypeOne(state *stateBenOr) int {
	return getTypeOne(state, 0) + getTypeOne(state, 1)
}

// returns number of type 2 messages with decidedness 'd' and value 'val'
func getTypeTwo(state *stateBenOr, d int, val byte) int {
	if d == 0 && val == 0 {
		return state.TwosU0[state.RecentRound]
	} else if d == 0 && val == 1 {
		return state.TwosU1[state.RecentRound]
	} else if d == 1 && val == 0 {
		return state.TwosD0[state.RecentRound]
	} else {
		return state.TwosD1[state.RecentRound]
	}
}

func getAllTypeTwo(state *stateBenOr) int {
	return getTypeTwo(state, msgUndecided, 0) + getTypeTwo(state, msgUndecided, 1) + getTypeTwo(state, msgDecided, 0) + getTypeTwo(state, msgDecided, 1)
}

func finaliseRound(node lib.Node, state *stateBenOr, decided bool) {
	r := state.RecentRound

	delete(state.Ones0, r)
	delete(state.Ones1, r)
	delete(state.ReceivedOnes, r)

	delete(state.TwosU0, r)
	delete(state.TwosU1, r)
	delete(state.TwosD0, r)
	delete(state.TwosD1, r)
	delete(state.ReceivedTwos, r)

	if decided {
		// participate in the next round
		sendAllBenOr(node, messageBenOr{Type: msgTypeOne, Round: r + 1, Value: state.EndingBit})
		sendAllBenOr(node, messageBenOr{Type: msgTypeTwo, Round: r + 1, Value: state.EndingBit, Decided: msgDecided})

		// let faulty processes know
		for i := 0; i < node.GetOutChannelsCount(); i++ {
			go node.SendMessage(i, nil)
		}
	}
}

func receiveAndProcessAnyMessage(node lib.Node, state *stateBenOr) {
	from, message := receiveAnyBenOr(node)
	if message.Round < state.RecentRound || (message.Decided != msgDecided && message.Decided != msgUndecided) || (message.Value != 0 && message.Value != 1) {
		return
	}

	if message.Type == msgTypeOne {
		// check if already sent
		if _, contains := state.ReceivedOnes[message.Round]; contains {
			if _, contains2 := state.ReceivedOnes[message.Round][from]; contains2 {
				// neighbour already sent type 1 message this round
				return
			}
		} else {
			state.ReceivedOnes[message.Round] = make(map[int]bool)
		}
		state.ReceivedOnes[message.Round][from] = true

		if message.Value == 0 {
			state.Ones0[message.Round]++
		} else {
			state.Ones1[message.Round]++
		}
	} else if message.Type == msgTypeTwo {
		if _, contains := state.ReceivedTwos[message.Round]; contains {
			if _, contains2 := state.ReceivedTwos[message.Round][from]; contains2 {
				// neighbour already sent type 2 message this round
				return
			}
		} else {
			state.ReceivedTwos[message.Round] = make(map[int]bool)
		}
		state.ReceivedTwos[message.Round][from] = true

		if message.Decided == 0 && message.Value == 0 {
			state.TwosU0[message.Round]++
		} else if message.Decided == 0 && message.Value == 1 {
			state.TwosU1[message.Round]++
		} else if message.Decided == 1 && message.Value == 0 {
			state.TwosD0[message.Round]++
		} else {
			state.TwosD1[message.Round]++
		}
	}
}

func processRound(node lib.Node) bool {
	state := getState(node)
	state.RecentRound++

	decided := false

	// send initial message to everyone
	sendAllBenOr(node, messageBenOr{
		Type:  msgTypeOne,
		Round: state.RecentRound,
		Value: state.EndingBit,
	})

	for getAllTypeOne(state) < state.N-state.T {
		receiveAndProcessAnyMessage(node, state)
	}

	var decideMessage messageBenOr
	if getTypeOne(state, 0) > (state.N+state.T)/2 {
		decideMessage = messageBenOr{Type: msgTypeTwo, Round: state.RecentRound, Value: 0, Decided: msgDecided}
	} else if getTypeOne(state, 1) > (state.N+state.T)/2 {
		decideMessage = messageBenOr{Type: msgTypeTwo, Round: state.RecentRound, Value: 1, Decided: msgDecided}
	} else {
		decideMessage = messageBenOr{Type: msgTypeTwo, Round: state.RecentRound, Decided: msgUndecided}
	}
	sendAllBenOr(node, decideMessage)

	for getAllTypeTwo(state) < state.N-state.T {
		receiveAndProcessAnyMessage(node, state)
	}

	zeroLen := getTypeTwo(state, msgDecided, 0)
	oneLen := getTypeTwo(state, msgDecided, 1)
	if zeroLen >= state.T+1 || oneLen >= state.T+1 {
		if zeroLen >= state.T+1 {
			state.EndingBit = 0
		} else if oneLen >= state.T+1 {
			state.EndingBit = 1
		}
		if zeroLen+oneLen > (state.N+state.T)/2 {
			decided = true
		}
	} else {
		state.EndingBit = byte(rand.Int() % 2)
	}

	finaliseRound(node, state, decided)
	setState(node, state)
	return decided
}

func initialiseBenOr(node lib.Node, startingBit byte, n, t int) {
	setState(node, &stateBenOr{
		StartingBit:  startingBit,
		EndingBit:    startingBit,
		N:            n,
		T:            t,
		ReceivedOnes: make(map[int]map[int]bool),
		Ones0:        map[int]int{},
		Ones1:        map[int]int{},
		ReceivedTwos: make(map[int]map[int]bool),
		TwosU0:       map[int]int{},
		TwosU1:       map[int]int{},
		TwosD0:       map[int]int{},
		TwosD1:       map[int]int{},
	})
}

func runBenOr(node lib.Node, startingBit byte, n, t int) {
	node.StartProcessing()

	initialiseBenOr(node, startingBit, n, t)

	for {
		if processRound(node) {
			break
		}
	}

	node.FinishProcessing(true)
}

func RunBenOr(processes []byte, behaviours []func(r int) byte) {
	t := len(behaviours)
	n := len(processes) + t

	nodes, runner := lib.BuildAsyncCompleteGraphWithLoops(n)

	for i := 0; i < n-t; i++ {
		startingBit := processes[i]
		if processes[i] != 0 && processes[i] != 1 {
			panic("Incorrect processes[] value")
		}

		go runBenOr(nodes[i], startingBit, n, t)
	}
	for i := 0; i < t; i++ {
		go runFaulty(behaviours[i], nodes[i+n-t], n, t)
	}

	runner.Run()
	runner.GetStats()

	checkBenOr(nodes, n, t)
}

// functions for faulty processes
func processMessageFaulty(behaviour func(r int) byte, node lib.Node, target int, msg messageBenOr) {
	responseValue := behaviour(msg.Round)
	responseMessage := messageBenOr{Type: msg.Type, Decided: msg.Decided, Round: msg.Round, Value: responseValue}

	newBytes, _ := json.Marshal(responseMessage)
	node.SendMessage(target, newBytes)
}

func runFaulty(behaviour func(r int) byte, node lib.Node, n, t int) {
	closed := 0

	node.StartProcessing()
	// accept messages until every nonfaulty process has finished
	for closed < n-t {
		from, bytes := node.ReceiveAnyMessage()
		if bytes == nil {
			closed++
			continue
		}
		var msg messageBenOr
		json.Unmarshal(bytes, &msg)

		go processMessageFaulty(behaviour, node, from, msg)
	}
	node.FinishProcessing(true)
}

func checkBenOr(nodes []lib.Node, n, t int) {
	if len(nodes) == 0 {
		log.Println("OK, length 0")
		return
	}
	firstBit := getState(nodes[0]).EndingBit

	// unanimous decision
	for i := 0; i < n-t; i++ {
		node := nodes[i]
		if decided := getState(node).EndingBit; decided != firstBit {
			panic("Decided values differ")
		}
	}
	// nontrivial
	for i := 0; i < n-t; i++ {
		node := nodes[i]
		if starting := getState(node).StartingBit; starting == firstBit {
			log.Println("OK, decided value:", firstBit)
			return
		}
	}

	panic("The decided value differs from every starting value")
}

func getState(node lib.Node) *stateBenOr {
	var result stateBenOr
	json.Unmarshal(node.GetState(), &result)
	return &result
}

func setState(node lib.Node, state *stateBenOr) {
	bytes, _ := json.Marshal(state)
	node.SetState(bytes)
}

func sendAllBenOr(node lib.Node, msg messageBenOr) {
	bytes, _ := json.Marshal(msg)
	channelsCount := node.GetOutChannelsCount()
	for i := 0; i < channelsCount; i++ {
		go node.SendMessage(i, bytes)
	}
}

func receiveAnyBenOr(node lib.Node) (int, messageBenOr) {
	// ignore nil & incorrectly formed messages
	for {
		from, bytes := node.ReceiveAnyMessage()
		if bytes == nil {
			continue
		}
		var msg messageBenOr
		err := json.Unmarshal(bytes, &msg)
		if err == nil {
			return from, msg
		}
	}
}
