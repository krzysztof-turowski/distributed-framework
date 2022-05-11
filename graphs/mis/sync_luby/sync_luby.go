package sync_luby

import (
	"encoding/json"
	"fmt"
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"log"
	"math/rand"
)

type statusType int

const (
	unknown statusType = iota
	winner
	loser
)

type state struct {
	Status             statusType
	Value              int
	RemainingNeighbors []bool
}

type message struct {
	Value int
}

func setState(v lib.Node, s state) {
	data, _ := json.Marshal(s)
	v.SetState(data)
}

func getState(v lib.Node) state {
	data := v.GetState()
	var s state
	s.RemainingNeighbors = make([]bool, v.GetInChannelsCount())
	json.Unmarshal(data, &s)
	return s
}

func sendMessage(v lib.Node, target int, msg message) {
	data, _ := json.Marshal(msg)
	v.SendMessage(target, data)
}

func sendNullMessage(v lib.Node, target int) {
	v.SendMessage(target, nil)
}

func receiveMessage(v lib.Node, source int) (message, bool) {
	data := v.ReceiveMessage(source)
	if data == nil {
		return message{}, false
	}
	var msg message
	json.Unmarshal(data, &msg)
	return msg, true
}

func initialize(v lib.Node) bool {
	neighbors := make([]bool, v.GetInChannelsCount())
	for i := range neighbors {
		neighbors[i] = true
	}
	setState(v, state{Status: unknown, RemainingNeighbors: neighbors})
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		sendNullMessage(v, i)
	}
	return false
}

func process(v lib.Node, round int) bool {
	state := getState(v)

	// react to messages from previous round
	switch (round + 2) % 3 {
	case 0: // after value-broadcast phase
		highestValue := true
		for i := 0; i < v.GetInChannelsCount(); i++ {
			if state.RemainingNeighbors[i] {
				msg, _ := receiveMessage(v, i)
				if msg.Value >= state.Value {
					highestValue = false
				}
			}
		}
		if highestValue {
			state.Status = winner
		}
	case 1: // after winner-broadcast phase
		for i := 0; i < v.GetInChannelsCount(); i++ {
			if state.RemainingNeighbors[i] {
				_, real := receiveMessage(v, i)
				if real {
					state.Status = loser
				}
			}
		}
	case 2: // after loser-broadcast phase
		for i := 0; i < v.GetInChannelsCount(); i++ {
			if state.RemainingNeighbors[i] {
				_, real := receiveMessage(v, i)
				if real {
					state.RemainingNeighbors[i] = false
				}
			}
		}
		if state.Status != unknown {
			setState(v, state)
			return true
		}
	}

	// send new messages
	switch round % 3 {
	case 0: // value-broadcast phase
		n := v.GetSize()
		state.Value = rand.Intn(n * n * n * n)
		for i := 0; i < v.GetOutChannelsCount(); i++ {
			if state.RemainingNeighbors[i] {
				sendMessage(v, i, message{Value: state.Value})
			}
		}
	case 1: // winner-broadcast phase
		for i := 0; i < v.GetOutChannelsCount(); i++ {
			if state.RemainingNeighbors[i] {
				if state.Status == winner {
					sendMessage(v, i, message{})
				} else {
					sendNullMessage(v, i)
				}
			}
		}
	case 2: // loser-broadcast phase
		for i := 0; i < v.GetOutChannelsCount(); i++ {
			if state.RemainingNeighbors[i] {
				if state.Status == loser {
					sendMessage(v, i, message{})
				} else {
					sendNullMessage(v, i)
				}
			}
		}
	}

	setState(v, state)
	return false
}

func run(v lib.Node) {
	v.StartProcessing()
	finish := initialize(v)
	v.FinishProcessing(finish)

	for round := 0; !finish; round++ {
		v.StartProcessing()
		finish = process(v, round)
		v.FinishProcessing(finish)
	}
}

func check(vertices []lib.Node) {
	// check statuses and neighborhoods
	for _, v := range vertices {
		state := getState(v)
		if state.Status != winner && state.Status != loser {
			panic(fmt.Sprint("Node ", v.GetIndex(), " has incorrect status"))
		}

		losers, winners := 0, 0
		for _, neighbor := range v.GetInNeighbors() {
			if getState(neighbor).Status == loser {
				losers++
			} else {
				winners++
			}
		}

		if state.Status == winner && winners > 0 {
			panic(fmt.Sprint("Node ", v.GetIndex(), " won together with its neighbor"))
		}
		if state.Status == loser && winners == 0 {
			panic(fmt.Sprint("Node ", v.GetIndex(), " lost together with all its neighbors"))
		}
	}
}

func Run(n int, p float64) (int, int) {
	vertices, synchronizer := lib.BuildSynchronizedRandomGraph(n, p)
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go run(v)
	}

	synchronizer.Synchronize(0)
	check(vertices)
	return synchronizer.GetStats()
}
