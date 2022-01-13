package ben_or

import (
	"encoding/json"
	"log"

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
	Val     byte
	Decided int
}

type vertexBenOr struct {
	node lib.Node

	// round -> [ids that have already sent this round]
	receivedOnes map[int]map[int]bool
	// ones[round][0/1] -> # of messages from round 'r' with value 0 or 1
	ones map[int]*[2]int

	// round -> [ids that have already sent this round]
	receivedTwos map[int]map[int]bool
	// twos[round][msgUndecided/msgDecided][0/1] -> # of messages from round 'r' which are decided/undecided, with value 0 or 1
	twos map[int]*[2][2]int

	// used to discard messages coming from previous rounds
	recentRound int
}

func (v *vertexBenOr) sendAllBenOr(msg messageBenOr) {
	bytes, _ := json.Marshal(msg)
	channelsCount := v.node.GetOutChannelsCount()
	for i := 0; i < channelsCount; i++ {
		go v.node.SendMessage(i, bytes)
	}
}

func (v *vertexBenOr) receiveAnyBenOr() (int, messageBenOr) {
	// ignore nil & incorrectly formed messages
	for {
		from, bytes := v.node.ReceiveAnyMessage()
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

func (v *vertexBenOr) processMessage(from int, msg messageBenOr) {
	if msg.Round < v.recentRound || (msg.Decided != msgDecided && msg.Decided != msgUndecided) || (msg.Val != 0 && msg.Val != 1) {
		return
	}

	if msg.Type == msgTypeOne {
		// check if already sent
		if _, contains := v.receivedOnes[msg.Round]; contains {
			if _, contains2 := v.receivedOnes[msg.Round][from]; contains2 {
				// neighbour already sent type 1 message this round
				return
			}
		} else {
			v.receivedOnes[msg.Round] = make(map[int]bool)
		}
		v.receivedOnes[msg.Round][from] = true

		if _, contains := v.ones[msg.Round]; !contains {
			v.ones[msg.Round] = &[2]int{0, 0}
		}

		(*v.ones[msg.Round])[msg.Val]++
	} else if msg.Type == msgTypeTwo {
		if _, contains := v.receivedTwos[msg.Round]; contains {
			if _, contains2 := v.receivedTwos[msg.Round][from]; contains2 {
				// neighbour already sent type 2 message this round
				return
			}
		} else {
			v.receivedTwos[msg.Round] = make(map[int]bool)
		}
		v.receivedTwos[msg.Round][from] = true

		if _, contains := v.twos[msg.Round]; !contains {
			v.twos[msg.Round] = &[2][2]int{{0, 0}, {0, 0}}
		}
		(*v.twos[msg.Round])[msg.Decided][msg.Val]++
	}
}

func (v *vertexBenOr) receiveAndProcessAnyMessage() {
	v.processMessage(v.receiveAnyBenOr())
}

func (v *vertexBenOr) initialiseBenOr() {
	v.receivedOnes = make(map[int]map[int]bool)
	v.receivedTwos = make(map[int]map[int]bool)
}

// returns number of type 1 messages from round 'r' with value 'x'
func (v *vertexBenOr) getLenTypeOne(r int, x byte) int {
	return v.ones[r][x]
}

// returns number of type 2 messages from round 'r' with value 'x' and decidedness 'd'
func (v *vertexBenOr) getLenTypeTwo(r int, d int, x byte) int {
	return v.twos[r][d][x]
}

func (v *vertexBenOr) initRound(r int) {
	for i := 0; i < v.node.GetInChannelsCount(); i++ {
		if _, contains := v.ones[r]; !contains {
			v.ones[r] = &[2]int{0, 0}
		}

		if _, contains := v.twos[r]; !contains {
			v.twos[r] = &[2][2]int{{0, 0}, {0, 0}}
		}
	}
}

// process one local round
func (v *vertexBenOr) processRound(r int, x byte, rand lib.Generator, n, t int) (bool, byte) {
	decided := false
	resultBit := x

	v.recentRound = r
	v.initRound(r)

	initMsg := messageBenOr{Type: msgTypeOne, Round: r, Val: resultBit}
	v.sendAllBenOr(initMsg)

	for v.getLenTypeOne(r, 0)+v.getLenTypeOne(r, 1) < n-t {
		v.receiveAndProcessAnyMessage()
	}

	var decideMessage messageBenOr
	if v.getLenTypeOne(r, 0) > (n+t)/2 {
		decideMessage = messageBenOr{Type: msgTypeTwo, Round: r, Val: 0, Decided: msgDecided}
	} else if v.getLenTypeOne(r, 1) > (n+t)/2 {
		decideMessage = messageBenOr{Type: msgTypeTwo, Round: r, Val: 1, Decided: msgDecided}
	} else {
		decideMessage = messageBenOr{Type: msgTypeTwo, Round: r, Decided: msgUndecided}
	}
	v.sendAllBenOr(decideMessage)

	for v.getLenTypeTwo(r, msgUndecided, 0)+v.getLenTypeTwo(r, msgUndecided, 1)+v.getLenTypeTwo(r, msgDecided, 0)+v.getLenTypeTwo(r, msgDecided, 1) < n-t {
		v.receiveAndProcessAnyMessage()
	}

	zeroLen := v.getLenTypeTwo(r, msgDecided, 0)
	oneLen := v.getLenTypeTwo(r, msgDecided, 1)
	if zeroLen >= t+1 || oneLen >= t+1 {
		if zeroLen >= t+1 {
			resultBit = 0
		} else if oneLen >= t+1 {
			resultBit = 1
		}
		if zeroLen+oneLen > (n+t)/2 {
			decided = true
		}
	} else {
		resultBit = byte(rand.Int() % 2)
	}

	// clean memory after finishing the round
	delete(v.ones, r)
	delete(v.receivedOnes, r)

	delete(v.twos, r)
	delete(v.receivedTwos, r)

	return decided, resultBit
}

func (v *vertexBenOr) runBenOr(startingBit byte, n, t int, rand lib.Generator) byte {
	v.node.StartProcessing()

	v.initialiseBenOr()

	x := startingBit
	decided := false

	r := 1
	for ; !decided; r++ {
		decided, x = v.processRound(r, x, rand, n, t)
	}

	v.sendAllBenOr(messageBenOr{Type: msgTypeOne, Round: r, Val: x})
	v.sendAllBenOr(messageBenOr{Type: msgTypeTwo, Round: r, Val: x, Decided: msgDecided})

	// let other processes know -> info only for faulty
	for i := 0; i < v.node.GetOutChannelsCount(); i++ {
		go v.node.SendMessage(i, nil)
	}

	// set resulting state
	v.node.GetState()[1] = x
	v.node.FinishProcessing(true)

	return x
}

func RunBenOr(processes []byte, behaviours []func(r int) byte) {
	t := len(behaviours)
	n := len(processes) + t

	nodes, runner := lib.BuildAsyncCompleteGraphWithLoops(n)

	vertices := make([]*vertexBenOr, 0)

	generator := lib.GetRandomGenerator()

	for i := 0; i < n-t; i++ {
		startingBit := processes[i]
		if processes[i] != 0 && processes[i] != 1 {
			panic("Incorrect processes[] value")
		}

		// set starting state, starting with (startingBit, ?), should end with (startingBit, decidedBit)
		nodes[i].SetState([]byte{startingBit, 2})

		v := vertexBenOr{node: nodes[i], ones: make(map[int]*[2]int), twos: make(map[int]*[2][2]int)}
		vertices = append(vertices, &v)
		go v.runBenOr(startingBit, n, t, generator)
	}
	for i := 0; i < t; i++ {
		go runFaulty(behaviours[i], nodes[i+n-t], n, t)
	}

	runner.Run()
	runner.GetStats()

	checkBenOr(vertices)
}

// functions for faulty processes
func processMessageFaulty(behaviour func(r int) byte, node lib.Node, target int, msg messageBenOr) {
	responseVal := behaviour(msg.Round)
	responseMessage := messageBenOr{Type: msg.Type, Decided: msg.Decided, Round: msg.Round, Val: responseVal}

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

// validation
func checkBenOr(vertices []*vertexBenOr) {
	if len(vertices) == 0 {
		log.Println("OK, length 0")
		return
	}
	firstBit := vertices[0].node.GetState()[1]

	// unanimous decision
	for _, v := range vertices {
		if decided := v.node.GetState()[1]; decided != firstBit {
			panic("Decided values differ")
		}
	}
	// nontrivial
	for _, v := range vertices {
		if starting := v.node.GetState()[0]; starting == firstBit {
			log.Println("OK, decided value:", firstBit)
			return
		}
	}

	panic("The decided value differs from every starting value")
}
