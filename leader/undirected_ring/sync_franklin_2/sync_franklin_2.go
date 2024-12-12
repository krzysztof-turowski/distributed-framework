package sync_franklin_2

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

const (
	any   int = -1
	left  int = 0
	right int = 1
)

const (
	unknown   byte = 0
	notLeader byte = 1
	leader    byte = 2
)

type state struct {
	State byte
	Stop  [2]bool
}

type message struct {
	Value int
	Stop  bool
}

func send(v lib.Node, direction int, msg message) {
	bytes, _ := json.Marshal(msg)
	v.SendMessage(direction, bytes)
}

func sendBoth(v lib.Node, msg message) {
	send(v, left, msg)
	send(v, right, msg)
}

func receive(v lib.Node, d int) (int, message) {
	var bytes []byte
	if d == any {
		d, bytes = v.ReceiveAnyMessage()
	} else {
		bytes = v.ReceiveMessage(d)
	}

	var msg message
	json.Unmarshal(bytes, &msg)
	if msg.Stop {
		s := getState(v)
		s.Stop[d] = true
		setState(v, s)
	}

	return d, msg
}

func receiveBoth(v lib.Node) (message, message) {
	var messages [2]message

	d, msg := receive(v, any)
	messages[d] = msg

	d, msg = receive(v, 1-d)
	messages[d] = msg

	return messages[0], messages[1]
}

func setState(v lib.Node, s state) {
	bytes, _ := json.Marshal(s)
	v.SetState(bytes)
}

func getState(v lib.Node) state {
	var s state
	json.Unmarshal(v.GetState(), &s)
	return s
}

func initialize(v lib.Node) bool {
	setState(v, state{})

	idx := v.GetIndex()
	sendBoth(v, message{Value: idx})

	return false
}

func processActive(v lib.Node) bool {
	index := v.GetIndex()
	msgLeft, msgRight := receiveBoth(v)
	msgMax := max(msgLeft.Value, msgRight.Value)

	if msgMax == index {
		setState(v, state{State: leader})
		sendBoth(v, message{index, true})
	} else if msgMax > index {
		setState(v, state{State: notLeader})
	} else {
		sendBoth(v, message{Value: index})
	}

	return false
}

func processRelay(v lib.Node) bool {
	d, msg1 := receive(v, any)
	send(v, 1-d, msg1)

	_, msg2 := receive(v, 1-d)
	send(v, d, msg2)

	return msg1.Stop || msg2.Stop
}

func processLeader(v lib.Node) bool {
	receiveBoth(v)

	return true
}

func process(v lib.Node) bool {
	s := getState(v)

	if s.State == unknown {
		return processActive(v)
	} else if s.State == notLeader {
		return processRelay(v)
	} else {
		return processLeader(v)
	}
}

func run(v lib.Node) {
	v.StartProcessing()
	finish := initialize(v)
	v.FinishProcessing(finish)

	for round := 1; !finish; round++ {
		v.StartProcessing()
		finish = process(v)
		v.FinishProcessing(finish)
	}
}

func check(vertices []lib.Node) {
	var leaderNode lib.Node
	var maxIndex int

	for _, v := range vertices {
		s := getState(v)
		if s.State == leader {
			if leaderNode != nil {
				panic(fmt.Sprint("Multiple leaders on the undirected ring: ",
					v.GetIndex(), leaderNode.GetIndex()))
			}
			leaderNode = v
		} else if s.State != notLeader {
			panic(fmt.Sprint("State is unknown upon end: ", v.GetIndex()))
		}

		maxIndex = max(maxIndex, v.GetIndex())
	}

	if leaderNode == nil {
		panic("There is no leader on the undirected ring")
	}

	if leaderNode.GetIndex() != maxIndex {
		panic(fmt.Sprint("Leader has value ", leaderNode.GetIndex(), " but max is ", maxIndex))
	}

	for _, v := range vertices {
		msgL := v.ReceiveMessageIfAvailable(left)
		msgR := v.ReceiveMessageIfAvailable(right)
		if len(msgL) > 0 || len(msgR) > 0 {
			panic(fmt.Sprint("Channel is not empty for vertex: ", v.GetIndex()))

		}
	}
}

func Run(n int) (int, int) {
	vertices, synchronizer := lib.BuildSynchronizedRing(n)
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go run(v)
	}
	synchronizer.Synchronize(0)
	check(vertices)

	return synchronizer.GetStats()
}
