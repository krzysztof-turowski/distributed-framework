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
	looser
)

type stateLubyMIS struct {
	Status  statusType
	Val     int
	RemNbrs []bool
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
	s.RemNbrs = make([]bool, v.GetInChannelsCount())
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
	nbrs := make([]bool, v.GetInChannelsCount())
	for i := range nbrs {
		nbrs[i] = true
	}
	setLubyMISState(v, stateLubyMIS{Status: unknown, RemNbrs: nbrs})
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
			if state.RemNbrs[i] {
				msg, _ := receiveLubyMISMessage(v, i)
				if msg.Value >= state.Val {
					highestValue = false
				}
			}
		}
		if highestValue {
			state.Status = winner
		}
	case 1: // after winner-broadcast phase
		for i := 0; i < v.GetInChannelsCount(); i++ {
			if state.RemNbrs[i] {
				_, real := receiveLubyMISMessage(v, i)
				if real {
					state.Status = looser
				}
			}
		}
	case 2: // after looser-broadcast phase
		for i := 0; i < v.GetInChannelsCount(); i++ {
			if state.RemNbrs[i] {
				_, real := receiveLubyMISMessage(v, i)
				if real {
					state.RemNbrs[i] = false
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
		state.Val = rand.Intn(n * n * n * n)
		for i := 0; i < v.GetOutChannelsCount(); i++ {
			if state.RemNbrs[i] {
				sendLubyMISMessage(v, i, messageLubyMIS{Value: state.Val})
			}
		}
	case 1: // winner-broadcast phase
		for i := 0; i < v.GetOutChannelsCount(); i++ {
			if state.RemNbrs[i] {
				if state.Status == winner {
					sendLubyMISMessage(v, i, messageLubyMIS{})
				} else {
					sendLubyMISNullMessage(v, i)
				}
			}
		}
	case 2: // looser-broadcast phase
		for i := 0; i < v.GetOutChannelsCount(); i++ {
			if state.RemNbrs[i] {
				if state.Status == looser {
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
	// build indices mapping
	indexToID := make(map[int]int)
	for i, v := range vertices {
		indexToID[v.GetIndex()] = i
	}

	// check statuses and neighborhoods
	for _, v := range vertices {
		state := getLubyMISState(v)
		if state.Status != winner && state.Status != looser {
			panic(fmt.Sprint("Node ", v.GetIndex(), " has incorrect status"))
		}

		loosers, winners := 0, 0
		for i := 0; i < v.GetInChannelsCount(); i++ {
			nei := vertices[indexToID[v.GetInNeighborIndex(i)]]
			neiState := getLubyMISState(nei)
			if neiState.Status == looser {
				loosers++
			} else {
				winners++
			}
		}

		if state.Status == winner && winners > 0 {
			panic(fmt.Sprint("Node ", v.GetIndex(), " won together with its neighbor"))
		}
		if state.Status == looser && winners == 0 {
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
