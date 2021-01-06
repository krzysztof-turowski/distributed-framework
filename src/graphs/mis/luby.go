package mis

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

type stateLuby struct {
	Status             statusType
	Value              int
	RemainingNeighbors []bool
}

type messageLuby struct {
	Value int
}

func setLubyState(v lib.Node, s stateLuby) {
	data, _ := json.Marshal(s)
	v.SetState(data)
}

func getLubyState(v lib.Node) stateLuby {
	data := v.GetState()
	var s stateLuby
	s.RemainingNeighbors = make([]bool, v.GetInChannelsCount())
	json.Unmarshal(data, &s)
	return s
}

func sendLubyMessage(v lib.Node, target int, msg messageLuby) {
	data, _ := json.Marshal(msg)
	v.SendMessage(target, data)
}

func sendLubyNullMessage(v lib.Node, target int) {
	v.SendMessage(target, nil)
}

func receiveLubyMessage(v lib.Node, source int) (messageLuby, bool) {
	data := v.ReceiveMessage(source)
	if data == nil {
		return messageLuby{}, false
	}
	var msg messageLuby
	json.Unmarshal(data, &msg)
	return msg, true
}

func initializeLuby(v lib.Node) bool {
	neighbors := make([]bool, v.GetInChannelsCount())
	for i := range neighbors {
		neighbors[i] = true
	}
	setLubyState(v, stateLuby{Status: unknown, RemainingNeighbors: neighbors})
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		sendLubyNullMessage(v, i)
	}
	return false
}

func processLuby(v lib.Node, round int) bool {
	state := getLubyState(v)

	// react to messages from previous round
	switch (round + 2) % 3 {
	case 0: // after value-broadcast phase
		highestValue := true
		for i := 0; i < v.GetInChannelsCount(); i++ {
			if state.RemainingNeighbors[i] {
				msg, _ := receiveLubyMessage(v, i)
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
				_, real := receiveLubyMessage(v, i)
				if real {
					state.Status = loser
				}
			}
		}
	case 2: // after loser-broadcast phase
		for i := 0; i < v.GetInChannelsCount(); i++ {
			if state.RemainingNeighbors[i] {
				_, real := receiveLubyMessage(v, i)
				if real {
					state.RemainingNeighbors[i] = false
				}
			}
		}
		if state.Status != unknown {
			setLubyState(v, state)
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
				sendLubyMessage(v, i, messageLuby{Value: state.Value})
			}
		}
	case 1: // winner-broadcast phase
		for i := 0; i < v.GetOutChannelsCount(); i++ {
			if state.RemainingNeighbors[i] {
				if state.Status == winner {
					sendLubyMessage(v, i, messageLuby{})
				} else {
					sendLubyNullMessage(v, i)
				}
			}
		}
	case 2: // loser-broadcast phase
		for i := 0; i < v.GetOutChannelsCount(); i++ {
			if state.RemainingNeighbors[i] {
				if state.Status == loser {
					sendLubyMessage(v, i, messageLuby{})
				} else {
					sendLubyNullMessage(v, i)
				}
			}
		}
	}

	setLubyState(v, state)
	return false
}

func runLuby(v lib.Node) {
	v.StartProcessing()
	finish := initializeLuby(v)
	v.FinishProcessing(finish)

	for round := 0; !finish; round++ {
		v.StartProcessing()
		finish = processLuby(v, round)
		v.FinishProcessing(finish)
	}
}

func checkLuby(vertices []lib.Node) {
	// check statuses and neighborhoods
	for _, v := range vertices {
		state := getLubyState(v)
		if state.Status != winner && state.Status != loser {
			panic(fmt.Sprint("Node ", v.GetIndex(), " has incorrect status"))
		}

		losers, winners := 0, 0
		for _, neighbor := range v.GetInNeighbors() {
			if getLubyState(neighbor).Status == loser {
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

func RunLuby(n int, p float64) (int, int) {
	vertices, synchronizer := lib.BuildSynchronizedRandomGraph(n, p)
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go runLuby(v)
	}

	synchronizer.Synchronize(0)
	checkLuby(vertices)
	return synchronizer.GetStats()
}
