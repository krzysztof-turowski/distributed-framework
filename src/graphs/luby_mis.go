package graphs

import (
	"encoding/json"
	"fmt"
	"lib"
	"log"
	"math/rand"
)

type statusType int

const (
	unknown statusType = iota
	winner
	loser
)

type stateLubyMIS struct {
	Status             statusType
	Value              int
	RemainingNeighbors []bool
}

type messageLubyMIS struct {
	Value int
}

func setLubyMISState(v lib.Node, s stateLubyMIS) {
	data, _ := json.Marshal(s)
	v.SetState(data)
}

func getLubyMISState(v lib.Node) stateLubyMIS {
	data := v.GetState()
	var s stateLubyMIS
	s.RemainingNeighbors = make([]bool, v.GetInChannelsCount())
	json.Unmarshal(data, &s)
	return s
}

func sendLubyMISMessage(v lib.Node, target int, msg messageLubyMIS) {
	data, _ := json.Marshal(msg)
	v.SendMessage(target, data)
}

func sendLubyMISNullMessage(v lib.Node, target int) {
	v.SendMessage(target, nil)
}

func receiveLubyMISMessage(v lib.Node, source int) (messageLubyMIS, bool) {
	data := v.ReceiveMessage(source)
	if data == nil {
		return messageLubyMIS{}, false
	}
	var msg messageLubyMIS
	json.Unmarshal(data, &msg)
	return msg, true
}

func initializeLubyMIS(v lib.Node) bool {
	neighbors := make([]bool, v.GetInChannelsCount())
	for i := range neighbors {
		neighbors[i] = true
	}
	setLubyMISState(v, stateLubyMIS{Status: unknown, RemainingNeighbors: neighbors})
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		sendLubyMISNullMessage(v, i)
	}
	return false
}

func processLubyMIS(v lib.Node, round int) bool {
	state := getLubyMISState(v)

	// react to messages from previous round
	switch (round + 2) % 3 {
	case 0: // after value-broadcast phase
		highestValue := true
		for i := 0; i < v.GetInChannelsCount(); i++ {
			if state.RemainingNeighbors[i] {
				msg, _ := receiveLubyMISMessage(v, i)
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
				_, real := receiveLubyMISMessage(v, i)
				if real {
					state.Status = loser
				}
			}
		}
	case 2: // after loser-broadcast phase
		for i := 0; i < v.GetInChannelsCount(); i++ {
			if state.RemainingNeighbors[i] {
				_, real := receiveLubyMISMessage(v, i)
				if real {
					state.RemainingNeighbors[i] = false
				}
			}
		}
		if state.Status != unknown {
			setLubyMISState(v, state)
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
				sendLubyMISMessage(v, i, messageLubyMIS{Value: state.Value})
			}
		}
	case 1: // winner-broadcast phase
		for i := 0; i < v.GetOutChannelsCount(); i++ {
			if state.RemainingNeighbors[i] {
				if state.Status == winner {
					sendLubyMISMessage(v, i, messageLubyMIS{})
				} else {
					sendLubyMISNullMessage(v, i)
				}
			}
		}
	case 2: // loser-broadcast phase
		for i := 0; i < v.GetOutChannelsCount(); i++ {
			if state.RemainingNeighbors[i] {
				if state.Status == loser {
					sendLubyMISMessage(v, i, messageLubyMIS{})
				} else {
					sendLubyMISNullMessage(v, i)
				}
			}
		}
	}

	setLubyMISState(v, state)
	return false
}

func runLubyMIS(v lib.Node) {
	v.StartProcessing()
	finish := initializeLubyMIS(v)
	v.FinishProcessing(finish)

	for round := 0; !finish; round++ {
		v.StartProcessing()
		finish = processLubyMIS(v, round)
		v.FinishProcessing(finish)
	}
}

func checkLubyMIS(vertices []lib.Node) {
	// check statuses and neighborhoods
	for _, v := range vertices {
		state := getLubyMISState(v)
		if state.Status != winner && state.Status != loser {
			panic(fmt.Sprint("Node ", v.GetIndex(), " has incorrect status"))
		}

		losers, winners := 0, 0
		for _, neighbor := range v.GetInNeighbors() {
			if getLubyMISState(neighbor).Status == loser {
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

func RunLubyMIS(n int, p float64) {
	vertices, synchronizer := lib.BuildSynchronizedRandomGraph(n, p)
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runLubyMIS(v)
	}

	synchronizer.Synchronize(0)
	synchronizer.GetStats()
	checkLubyMIS(vertices)
}
