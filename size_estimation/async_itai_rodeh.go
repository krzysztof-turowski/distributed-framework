package async_itai_rodeh

import (
	"encoding/json"
	"log"
	"math/rand"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

type State struct {
	Estimate   int
	Id         int
	Confidence int
	Terminated bool
}

type Message struct {
	Estimate         int
	Id               int
	HopCount         int
	TerminationRound bool
}

func max(a int, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func encode(object interface{}) []byte {
	data, _ := json.Marshal(object)
	return data
}

func decode(data []byte, object interface{}) {
	json.Unmarshal(data, &object)
}

func setState(v lib.Node, s *State) {
	v.SetState(encode(s))
}

func getState(v lib.Node) State {
	var s State
	decode(v.GetState(), &s)
	return s
}

func send(v lib.Node, estimate int, id int, hopCount int) {
	go v.SendMessage(0, encode(Message{
		Estimate:         estimate,
		Id:               id,
		HopCount:         hopCount,
		TerminationRound: false,
	}))
}

func sendTerminateMessage(v lib.Node) {
	go v.SendMessage(0, encode(Message{
		Estimate:         getState(v).Estimate,
		TerminationRound: true,
	}))
}

func receive(v lib.Node) Message {
	var m Message
	decode(v.ReceiveMessage(0), &m)
	return m
}

func process(v lib.Node) bool {
	s := getState(v)
	m := receive(v)
	defer setState(v, &s)

	if m.TerminationRound {
		s.Estimate = max(s.Estimate, m.Estimate)
		s.Terminated = true
		sendTerminateMessage(v)
		return false
	}

	if m.Estimate < s.Estimate {
		return true
	}

	if m.HopCount < m.Estimate {
		send(v, m.Estimate, m.Id, m.HopCount+1)
		if s.Estimate < m.Estimate {
			s.Estimate = m.Estimate
			s.Confidence = 0
		}
		return true
	}

	if m.Estimate > s.Estimate || m.Id != s.Id {
		s.Estimate = m.Estimate + 1
		s.Confidence = 0
		return false
	}

	s.Confidence++
	return false
}

func initialize(v lib.Node) {
	setState(v, &State{
		Estimate:   1,
		Confidence: 0,
		Terminated: false,
	})
}

func initiateTerminationRound(v lib.Node) {
	s := getState(v)
	defer setState(v, &s)

	sendTerminateMessage(v)
	for {
		if m := receive(v); m.TerminationRound {
			s.Estimate = max(s.Estimate, m.Estimate)
			return
		}
	}
}

func run(v lib.Node, r int) {
	v.StartProcessing()
	defer v.FinishProcessing(true)

	initialize(v)
	for {
		s := getState(v)
		if s.Terminated {
			return
		}
		if s.Confidence == r {
			initiateTerminationRound(v)
			return
		}
		s.Id = rand.Int()
		setState(v, &s)
		send(v, s.Estimate, s.Id, 1)
		for process(v) {
		}
	}
}

func checkEstimatesCorrect(vertices []lib.Node) (bool, float64) {
	correct := true
	sum := int64(0)

	for _, v := range vertices {
		estimate := getState(v).Estimate
		if estimate != len(vertices) {
			correct = false
		}
		sum += int64(estimate)
	}
	return correct, float64(sum) / float64(len(vertices))
}

func Run(n int, r int) (int, int) {
	vertices, runner := lib.BuildDirectedRing(n)

	for _, v := range vertices {
		log.Println("Running v", v.GetIndex())
		go run(v, r)
	}

	runner.Run(false)
	correct, averageEstimate := checkEstimatesCorrect(vertices)
	log.Println("All estimates correct: ", correct)
	log.Println("Average estimate: ", averageEstimate)

	return runner.GetStats()
}
