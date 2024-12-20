package async_chang_roberts

import (
	"encoding/json"
	"fmt"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
	"math/rand"
	"time"
)

const (
	LeaderSelected = iota
	LeaderUnknown  = iota
)

type message struct {
	Value int
	Mode  int
}

type finalState struct {
	IsLeader         bool
	LargestIndexSeen int
}

func process(v lib.Node) bool {
	largestIndexSeen := v.GetIndex()
	send(v.GetIndex(), LeaderUnknown, 0, v) // starting direction (0 or 1) is arbitrary
	for turn := 0; ; turn++ {
		cameFrom, message := receive(v)

		if message.Mode == LeaderUnknown {
			if v.GetIndex() == message.Value {
				send(v.GetIndex(), LeaderSelected, 0, v) // finishing message direction (0 or 1) is arbitrary
			} else {
				if message.Value > largestIndexSeen {
					largestIndexSeen = message.Value
					go send(largestIndexSeen, LeaderUnknown, 1-cameFrom, v)
				}
			}
		} else {
			if v.GetIndex() != largestIndexSeen {
				send(message.Value, LeaderSelected, 1-cameFrom, v)
			}
			setFinalState(v.GetIndex() == largestIndexSeen, largestIndexSeen, v)
			return v.GetIndex() == largestIndexSeen
		}
	}
}

func setFinalState(isLeader bool, largestIndexSeen int, v lib.Node) {
	msg, _ := json.Marshal(finalState{IsLeader: isLeader, LargestIndexSeen: largestIndexSeen})
	v.SetState(msg)
}

func send(value, mode, direction int, v lib.Node) {
	msg, _ := json.Marshal(message{Value: value, Mode: mode})
	v.SendMessage(direction, msg)
}

func receive(v lib.Node) (int, message) {
	var m message
	cameFrom, encodedMessage := v.ReceiveAnyMessage()
	json.Unmarshal(encodedMessage, &m)
	return cameFrom, m
}

func run(v lib.Node) {
	v.StartProcessing()
	process(v)
	v.FinishProcessing(true)
}

func Run(n int) (int, int) {
	rand.Seed(time.Now().UnixNano())

	vertices, runner := lib.BuildRing(n)

	for _, v := range vertices {
		log.Println("Running node", v.GetIndex())
		go run(v)
	}

	runner.Run(false)
	check(vertices)
	return runner.GetStats()
}

func check(vertices []lib.Node) {
	var leader lib.Node = nil
	var s finalState
	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if s.IsLeader {
			if leader != nil {
				panic("More than one leader selected")
			}
			leader = v
		}
	}
	if leader == nil {
		panic("No leader selected")
	}
	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if s.LargestIndexSeen != leader.GetIndex() {
			panic(fmt.Sprint(
				"Leader has index ", leader.GetIndex(), " but node ", v.GetIndex(),
				" has maximum index seen", s.LargestIndexSeen))
		}
	}
}
