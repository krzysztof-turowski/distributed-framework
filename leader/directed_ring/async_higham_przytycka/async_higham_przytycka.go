package async_higham_przytycka

import (
	"encoding/json"
	"math/rand"
	"time"

	. "github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/common"
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

type message struct {
	Id      int
	Round   int
	Counter int
}

type state struct {
	PrevMessage message
	Status      ModeType
}

func send(v lib.Node, m message) {
	outMessage, _ := json.Marshal(m)
	s := getState(v)
	s.PrevMessage = m
	setState(v, s)
	v.SendMessage(0, outMessage)
}

func getState(v lib.Node) state {
	var s state
	json.Unmarshal(v.GetState(), &s)
	return s
}

func setState(v lib.Node, s state) {
	data, _ := json.Marshal(s)
	v.SetState(data)
}

func receive(v lib.Node) (state, message) {
	var m message
	inMessage := v.ReceiveMessage(0)
	json.Unmarshal(inMessage, &m)
	return getState(v), m
}

func shouldDeleteMessage(newMessage message, oldMessage message) bool {
	if newMessage.Round == oldMessage.Round {
		if newMessage.Round%2 == 0 {
			return newMessage.Id < oldMessage.Id
		} else {
			return newMessage.Id > oldMessage.Id
		}
	} else {
		return false
	}
}

func shouldPromoteMessage(newMessage message, oldMessage message) bool {
	if newMessage.Round%2 == 0 {
		return oldMessage.Round+1 == newMessage.Round &&
			newMessage.Id > oldMessage.Id
	} else {
		return newMessage.Counter == 0 ||
			(oldMessage.Round == newMessage.Round && newMessage.Id > oldMessage.Id)
	}
}

func finalMessage() message {
	return message{Round: -1}
}

func isFinal(m message) bool {
	return m.Round == -1
}

func equals(newMessage message, oldMessage message) bool {
	return newMessage.Id == oldMessage.Id && newMessage.Round == oldMessage.Round
}

func fibonacci(n int) int {
	prev := 1
	cur := 0
	for i := 0; i < n; i++ {
		cur = prev + cur
		prev = prev - cur
	}
	return cur
}

func process(v lib.Node) bool {
	state, message := receive(v)
	if equals(message, state.PrevMessage) {
		state.Status = Leader
		setState(v, state)
		send(v, finalMessage())
		return false
	}

	if isFinal(message) {
		state.Status = NonLeader
		setState(v, state)
		send(v, finalMessage())
		return false
	}

	if !shouldDeleteMessage(message, state.PrevMessage) {
		if shouldPromoteMessage(message, state.PrevMessage) {
			message.Round++
			message.Counter = fibonacci(message.Round + 1)
		}
		message.Counter--
		send(v, message)
	}
	return true
}

func initialize(v lib.Node, id int) {
	send(v, message{id, 0, 1})
}

func run(v lib.Node, id int) {
	v.StartProcessing()
	initialize(v, id)
	for process(v) {
	}
	v.FinishProcessing(true)
}

func check(vertices []lib.Node) {
	var leadNode lib.Node
	for _, v := range vertices {
		s := getState(v)
		if s.Status == Leader {
			if leadNode != nil {
				panic("There are multiple leaders")
			}
			leadNode = v
		}
	}
	if leadNode == nil {
		panic("There is no Leader")
	}
}

func Run(n int) (int, int) {
	rand.Seed(time.Now().UnixNano())

	vertices, runner := lib.BuildDirectedRing(n)
	ids := rand.Perm(n)

	for i, v := range vertices {
		go run(v, ids[i])
	}

	runner.Run(false)
	check(vertices)
	return runner.GetStats()
}
