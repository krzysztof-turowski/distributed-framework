package async_syrotiuk_pachl_2

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
	UNDECIDED   = 2
)

type message struct {
	Distance int
	Vote     int
	Stop     bool
	Decided  bool
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
		go send(v, 1-from, msg)

		if !msg.Decided {
			setState(v, state{UNDECIDED})
		} else if from == LEFT {
			setState(v, state{NOT_FLIPPED})
		} else {
			setState(v, state{FLIPPED})
		}

		return true
	} else {
		if msg.Distance == v.GetSize() {
			if msg.Vote > 0 {
				go send(v, RIGHT, message{
					Stop:    true,
					Decided: true,
				})
				setState(v, state{NOT_FLIPPED})
			} else {
				go send(v, RIGHT, message{
					Stop:    true,
					Decided: false,
				})
				setState(v, state{UNDECIDED})
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
	undecidedNodes := 0
	for _, v := range vertices {
		state := getState(v)
		if state.State == UNDECIDED {
			undecidedNodes += 1
		}

		// Check if all states are set to some value
		if state.State == UNKNOWN {
			panic("Node state unknown at the end")
		}
	}
	if undecidedNodes != 0 {
		// Check if all nodes are undecided
		if undecidedNodes != len(vertices) {
			panic("Some nodes are oriented and some are not")
		}

		// Check if orientation can't be decided
		diff := 0
		v := vertices[0]
		prev := vertices[0]
		for {
			u0 := v.GetInNeighbors()[0]
			u1 := v.GetInNeighbors()[1]

			if u0 == prev {
				diff += 1
				prev = v
				v = u1
			} else {
				diff -= 1
				prev = v
				v = u0
			}

			if v == vertices[0] {
				break
			}
		}
		if diff != 0 {
			panic("Orientation undecided but majority agrees on some orientation")
		}
	}

	// Check if orientation is the same for all nodes
	for _, v := range vertices {
		u0 := v.GetInNeighbors()[0]
		u1 := v.GetInNeighbors()[1]

		stateV := getState(v)
		stateU0 := getState(u0)
		stateU1 := getState(u1)

		// if the states are the same, then v->u, u->v can't be both left (right) channels
		if ((stateV.State == stateU0.State) != (u0.GetInNeighbors()[1] == v)) ||
			((stateV.State == stateU1.State) != (u1.GetInNeighbors()[0] == v)) {
			panic("Some nodes are oriented differently")
		}
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
