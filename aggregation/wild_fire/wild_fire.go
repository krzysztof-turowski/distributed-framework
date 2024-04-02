package wildfire

import (
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

type pending struct {
	q    string
	D    int
	A    int
	from int
}

type state struct {
	Active bool
	D      int
	A      int
}

func (instance state) TimedOut(time int, node lib.Node) bool {
	if instance.D == 0 {
		// we haven't delivered the diameter extimation from the query node
		if time < 10 {
			return false
		} else {
			// after 10 rounds we declare the node DISCONNECTED
			log.Println("DISCONNECTED-FROM-THE-QUERY-NODE [", node.GetIndex(), "]")
			return true
		}

	}
	// t()-t0' > 2D'δ
	return time >= instance.D*2
}

func (instance state) SetOn(v lib.Node) {
	data, _ := json.Marshal(instance)
	v.SetState(data)
}

func StateOf(node lib.Node) state {
	data := node.GetState()
	var res state
	json.Unmarshal(data, &res)
	return res
}

type message struct {
	Query string
	D     int
	A     int
}

func (instance message) SendToNeighbor(node lib.Node, index int) {
	mess, _ := json.Marshal(instance)
	node.SendMessage(index, mess)
}

func (instance message) SendToAllNeighbors(node lib.Node) {
	mess, _ := json.Marshal(instance)
	for index := 0; index < node.GetInChannelsCount(); index++ {
		node.SendMessage(index, mess)
	}
}

func TakeMessageFrom(node lib.Node, index int) (message, bool) {
	mess := node.ReceiveMessageIfAvailable(index)
	if mess == nil {
		return message{}, false
	}

	var res message
	json.Unmarshal(mess, &res)
	return res, true
}

func combine(query string, val1 int, val2 int) int {
	if query == "min" {
		if val1 >= val2 {
			return val2
		} else {
			return val1
		}
	} else if query == "max" {
		if val1 >= val2 {
			return val1
		} else {
			return val2
		}
	} else {
		panic("unsupported query function: " + query)
	}
}

func processPendings(node lib.Node, pendings []pending) {
	state := StateOf(node)
	if !state.Active {
		state.D = pendings[0].D
	}

	processes := make(map[int]bool, node.GetInChannelsCount())
	for i := 0; i < node.GetInChannelsCount(); i++ {
		processes[i] = false
	}
	Anew := state.A
	for _, c := range pendings {
		processes[c.from] = true
		Anew = combine(c.q, c.A, Anew)
	}
	for _, c := range pendings {
		if c.A != Anew {
			log.Println("CONVERGECAST [", node.GetIndex(), "]--", Anew, "->[", node.GetOutNeighbors()[c.from].GetIndex(), "]")
			message{c.q, c.D, Anew}.SendToNeighbor(node, c.from)
		}
	}
	if Anew != state.A || !state.Active {
		state.A = Anew
		for from, processed := range processes {
			if !processed {
				log.Println("CONVERGECAST [", node.GetIndex(), "]--", Anew, "->[", node.GetOutNeighbors()[from].GetIndex(), "]")
				message{pendings[0].q, state.D, Anew}.SendToNeighbor(node, from)
			}
		}
	}

	state.Active = true
	state.SetOn(node)
}

func queryStep(node lib.Node, query string, diameterExtimation int, localValue int) bool {
	state{true, diameterExtimation, localValue}.SetOn(node)
	message{query, diameterExtimation, localValue}.SendToAllNeighbors(node)
	log.Println("BROADCAST [", node.GetIndex(), "]--", localValue, "->")
	return false
}

func initialize(node lib.Node, localValue int) bool {
	state{false, 0, localValue}.SetOn(node)
	return false
}

func process(node lib.Node, time int) bool {
	state := StateOf(node)
	if state.TimedOut(time, node) {
		return true
	}

	pendings := make([]pending, 0)
	for from := 0; from < node.GetInChannelsCount(); from++ {
		mess, present := TakeMessageFrom(node, from)
		if present {
			pendings = append(pendings, pending{mess.Query, mess.D, mess.A, from})
		}
	}

	if len(pendings) > 0 {
		processPendings(node, pendings)
	}

	return false
}

func queryNode(node lib.Node, query string, diameterExtimation int, localValue int) {
	node.StartProcessing()
	finish := queryStep(node, query, diameterExtimation, localValue)
	node.FinishProcessing(false)
	for time := 1; !finish; time++ {
		node.StartProcessing()
		finish := process(node, time)
		if finish {
			log.Println("QUERY-NODE-NATURAL-STOP [", node.GetIndex(), "]")
		}
		node.FinishProcessing(finish)
	}
	node.IgnoreFutureMessages()
}

func correctNode(node lib.Node, localValue int) {
	node.StartProcessing()
	finish := initialize(node, localValue)
	node.FinishProcessing(false)
	for time := 1; !finish; time++ {
		node.StartProcessing()
		finish := process(node, time)
		if finish {
			log.Println("CORRECT-NODE-NATURAL-STOP [", node.GetIndex(), "]")
		}
		node.FinishProcessing(finish)
	}
	node.IgnoreFutureMessages()
}

func faultyNode(node lib.Node, localValue int, faultyProbability float64) {
	node.StartProcessing()
	finish := initialize(node, localValue)
	node.FinishProcessing(false)
	for time := 1; !finish; time++ {
		node.StartProcessing()
		finish := process(node, time)
		if finish {
			log.Println("FAULTY-NODE-NATURAL-STOP [", node.GetIndex(), "]")
		} else if faultyProbability < rand.Float64() {
			finish = true
			log.Println("FAIL-STOP-CRASH [", node.GetIndex(), "]")
		}
		node.FinishProcessing(finish)
	}
	node.IgnoreFutureMessages()
}

func Run(n int, edgeProbability float64, faultyProbability float64, query string, diffValues int, diameterExtimation int) (int, int) {
	expectedFaulyProcesses := float64(n) * faultyProbability

	// set the channel size 3 times the value of expected number of neighbors
	// this value should prevent the fact that a node could be blocking on sending
	neighborsExpectedValue := float64(n) * edgeProbability
	notEdgeProbability := float64(1) - edgeProbability
	channelSize := int(math.Ceil(neighborsExpectedValue * float64(3)))

	log.Println("Wildfire protocol starts: n:", n, "edgeProbability", edgeProbability, " notEdgeProbability: ", notEdgeProbability, " neighborsExpectedValue: ", neighborsExpectedValue)
	log.Println("faultyProbability: ", faultyProbability, " expectedFaulyProcesses: ", expectedFaulyProcesses)
	log.Println("query: ", query, " diffValues: ", diffValues, " diameterExtimation: ", diameterExtimation)

	nodes, synchronizer := lib.BuildSynchronizedRandomBufferedChannelsGraph(n, notEdgeProbability, channelSize)
	for i, node := range nodes {
		localValue := 0
		if query == "min" {
			// with min we assign smaller values to the remote nodes
			localValue = int(float32(n-i) * float32(diffValues) / float32(n))
		} else if query == "max" {
			// with max we assign larger values to the remote nodes
			localValue = int(float32(i) * float32(diffValues) / float32(n))
		} else {
			panic("unsupported query function: " + query)
		}

		if i == 0 {
			log.Println("Query node", node.GetIndex(), "about to run, with local query value of: ", localValue)
			go queryNode(node, query, diameterExtimation, localValue)
		} else if faultyProbability < rand.Float64() {
			log.Println("Correct node", node.GetIndex(), "about to run, with local query value of: ", localValue)
			go correctNode(node, localValue)
		} else {
			log.Println("Faulty node", node.GetIndex(), "about to run, with local query value of: ", localValue)
			go faultyNode(node, localValue, faultyProbability)
		}
	}

	synchronizer.Synchronize(500 * time.Millisecond)
	messages, _ := synchronizer.GetStats()
	return StateOf(nodes[0]).A, messages
}
