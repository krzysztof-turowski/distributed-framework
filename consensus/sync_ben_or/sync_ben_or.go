package sync_ben_or

import (
	"encoding/json"
	"log"
	"math/rand"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

type stateBenOr struct {
	StartingBit byte
	EndingBit   byte
	Done        bool

	TypeOne  [2]int
	TypeTwoU [2]int
	TypeTwoD [2]int

	N int
	T int

	// keeps track of nil messages given round
	Finished int
}

type messageTypeBenOr int

const (
	messageTypeOne messageTypeBenOr = iota
	messageTypeTwo
)

type messageBenOr struct {
	Type    messageTypeBenOr
	Value   byte
	Decided bool

	AlreadyFinished bool
}

func processMessage(v lib.Node, message messageBenOr) {
	state := getState(v)
	if message.AlreadyFinished {
		state.Finished++
	}
	if message.Type == messageTypeOne {
		state.TypeOne[message.Value]++
	} else if message.Type == messageTypeTwo {
		if message.Decided {
			state.TypeTwoD[message.Value]++
		} else {
			state.TypeTwoU[message.Value]++
		}
	}
	setState(v, state)
}

func initializeRound(v lib.Node) {
	state := getState(v)
	setState(v, &stateBenOr{
		StartingBit: state.StartingBit,
		EndingBit:   state.EndingBit,
		Done:        state.Done,
		TypeOne:     [2]int{0, 0},
		TypeTwoU:    [2]int{0, 0},
		TypeTwoD:    [2]int{0, 0},
		N:           state.N,
		T:           state.T,
		Finished:    0,
	})
}

func processRound(v lib.Node) bool {
	initializeRound(v)
	state := getState(v)

	sendAll(v, messageBenOr{Type: messageTypeOne, Value: state.EndingBit, AlreadyFinished: state.Done})

	// wait for type 1 messages
	receiveAndProcessMessages(v)

	state = getState(v)

	var decideMessage messageBenOr
	if 2*state.TypeOne[0] > state.N+state.T {
		decideMessage = messageBenOr{Type: messageTypeTwo, Value: 0, Decided: true, AlreadyFinished: state.Done}
	} else if 2*state.TypeOne[1] > state.N+state.T {
		decideMessage = messageBenOr{Type: messageTypeTwo, Value: 1, Decided: true, AlreadyFinished: state.Done}
	} else {
		decideMessage = messageBenOr{Type: messageTypeTwo, Decided: false, AlreadyFinished: state.Done}
	}
	sendAll(v, decideMessage)

	// wait for type 2 messages
	receiveAndProcessMessages(v)
	state = getState(v)

	if zeros, ones := state.TypeTwoD[0], state.TypeTwoD[1]; zeros >= state.T+1 || ones >= state.T+1 {
		if zeros >= state.T+1 {
			state.EndingBit = 0
		} else if ones >= state.T+1 {
			state.EndingBit = 1
		}
		if 2*(zeros+ones) > state.N+state.T {
			state.Done = true
		}
	} else {
		state.EndingBit = byte(rand.Intn(2))
	}

	setState(v, state)
	return state.Finished == 2*(state.N-state.T)
}

func initialize(v lib.Node, startingBit byte, n, t int) bool {
	setState(v, &stateBenOr{
		StartingBit: startingBit,
		EndingBit:   startingBit,
		Done:        false,
		TypeOne:     [2]int{0, 0},
		TypeTwoU:    [2]int{0, 0},
		TypeTwoD:    [2]int{0, 0},
		N:           n,
		T:           t,
		Finished:    0,
	})
	return false
}

func run(v lib.Node, startingBit byte, n, t int) {
	v.StartProcessing()
	finish := initialize(v, startingBit, n, t)
	v.FinishProcessing(finish)

	for r := 1; !finish; r++ {
		v.StartProcessing()
		finish = processRound(v)
		v.FinishProcessing(finish)
	}
}

func Run(processes []byte, behaviours []func(r int) byte) (int, int) {
	t := len(behaviours)
	n := len(processes) + t
	if 5*t >= n {
		panic("t too big")
	}
	nodes, synchronizer := lib.BuildCompleteGraphWithLoops(n, false, lib.GetRandomGenerator())

	for i := 0; i < n-t; i++ {
		bit := processes[i]
		if bit != 0 && bit != 1 {
			panic("invalid processes[] value")
		}
		go run(nodes[i], bit, n, t)
	}

	for i := 0; i < t; i++ {
		go runFaulty(nodes[n-t+i], behaviours[i], n, t)
	}

	synchronizer.Synchronize(0)
	check(nodes, n, t)

	return synchronizer.GetStats()
}

func processFaultyRound(v lib.Node, r int, behaviour func(r int) byte) bool {
	initializeRound(v)

	sendAll(v, messageBenOr{Type: messageTypeOne, Value: behaviour(r)})
	receiveAndProcessMessages(v)

	sendAll(v, messageBenOr{Type: messageTypeTwo, Value: behaviour(r), Decided: true})
	receiveAndProcessMessages(v)

	state := getState(v)
	return state.Finished == 2*(state.N-state.T)
}

func runFaulty(v lib.Node, behaviour func(r int) byte, n, t int) {
	v.StartProcessing()
	finished := initialize(v, behaviour(0), n, t)
	v.FinishProcessing(finished)

	for r := 1; !finished; r++ {
		v.StartProcessing()
		finished = processFaultyRound(v, r, behaviour)
		v.FinishProcessing(finished)
	}
}

func check(nodes []lib.Node, n, t int) {
	if len(nodes) == 0 {
		log.Println("OK, length 0")
		return
	}

	// everyone decided
	for i := 0; i < n-t; i++ {
		if !getState(nodes[i]).Done {
			panic("Node did not finish")
		}
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

func sendAll(node lib.Node, message messageBenOr) {
	bytes, _ := json.Marshal(message)
	for i := 0; i < node.GetOutChannelsCount(); i++ {
		node.SendMessage(i, bytes)
	}
}

func receiveAndProcessMessages(node lib.Node) {
	for i := 0; i < node.GetInChannelsCount(); i++ {
		bytes := node.ReceiveMessage(i)
		var msg messageBenOr
		json.Unmarshal(bytes, &msg)
		processMessage(node, msg)
	}
}
