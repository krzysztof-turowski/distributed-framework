package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

const (
	ANY   = -1
	LEFT  = 0
	RIGHT = 1

	UNKNOWN     = 0
	FLIPPED     = -1
	NOT_FLIPPED = 1
)

type message struct {
	Distance int
	Vote     int
	Stop     bool
}

type state struct {
	State int
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

func send(v lib.Node, d int, msg message) {
	bytes, _ := json.Marshal(msg)
	v.SendMessage(d, bytes)
}

func receive(v lib.Node, d int) (int, message) {
	var bytes []byte
	if d == ANY {
		d, bytes = v.ReceiveAnyMessage()
	} else {
		bytes = v.ReceiveMessage(d)
	}

	var msg message
	json.Unmarshal(bytes, &msg)

	return d, msg
}

func initialize(v lib.Node) {
	setState(v, state{UNKNOWN})
	go send(v, RIGHT, message{
		Distance: 1,
		Vote:     1,
	})
}

func process(v lib.Node) bool {
	from, msg := receive(v, ANY)
	if msg.Stop {
		go send(v, 1-from, message{Stop: true})

		if from == LEFT {
			setState(v, state{NOT_FLIPPED})
		} else {
			setState(v, state{FLIPPED})
		}

		return true
	} else {
		if msg.Distance == v.GetSize() {
			if msg.Vote >= 0 {
				go send(v, RIGHT, message{Stop: true})
				setState(v, state{NOT_FLIPPED})
			} else {
				go send(v, LEFT, message{Stop: true})
				setState(v, state{FLIPPED})
			}

			return true
		} else if from == LEFT {
			go send(v, RIGHT, message{
				Vote:     msg.Vote + 1,
				Distance: msg.Distance + 1,
			})
		} else if from == RIGHT && msg.Vote > 0 {
			go send(v, LEFT, message{
				Vote:     msg.Vote - 1,
				Distance: msg.Distance + 1,
			})
		}
	}

	return false
}

func run(v lib.Node) {
	v.StartProcessing()

	initialize(v)
	for !process(v) {
	}

	v.FinishProcessing(true)
}

func check(vertices []lib.Node) {
	for _, v := range vertices {
		state := getState(v)
		println(v.GetIndex(), state.State)
	}
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

func main() {
	Run(10001)
}
