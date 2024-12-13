package async_higham_przytycka

import (
	"encoding/json"
	"log"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

// types and constants
const (
	LEFT  = 0
	RIGHT = 1
)

const (
	NORMAL    = 0
	BROADCAST = 1
)

type Msg struct {
	msg_type int
	id       int
	round    int
	counter  int
}

type state struct {
	leader int
}

//

func nullMsg() Msg {
	return Msg{
		msg_type: NORMAL,
		id:       -1,
		round:    -1,
		counter:  0,
	}
}

func firstMsg(id int) Msg {
	return Msg{
		msg_type: NORMAL,
		id:       id,
		round:    0,
		counter:  0,
	}
}

func Fib(n int) int {
	a := 0
	b := 1
	for i := 1; i < n; i++ {
		c := a + b
		a = b
		b = c
	}
	return b
}

func odd(i int) bool {
	return i%2 != 0
}

func even(i int) bool {
	return i%2 == 0
}

func leaderTest(prev_msg Msg, msg Msg) bool {
	return msg.id == prev_msg.id && msg.round == prev_msg.round
}

func casualtyTest(prev_msg Msg, msg Msg) bool {
	return prev_msg.round == msg.round && ((odd(msg.round) && msg.id > prev_msg.id) || (even(msg.round) && msg.id < prev_msg.id))
}

func promotionTest(prev_msg Msg, msg Msg) bool {
	return (even(msg.round) && prev_msg.round == msg.round-1 && msg.id > prev_msg.id) || (odd(msg.round) && (msg.counter == 0 || (prev_msg.round == msg.round && msg.id < prev_msg.id)))
}

func highamPrzytycka(proc_id int, rec func() Msg, send func(Msg), output chan<- int) {
	msg := firstMsg(proc_id)
	prev_msg := nullMsg()

	// Leader election
	for !leaderTest(prev_msg, msg) {
		if !casualtyTest(prev_msg, msg) {
			if promotionTest(prev_msg, msg) {
				msg.round += 1
				msg.counter = Fib(msg.round + 2)
			}

			prev_msg = msg

			msg.counter -= 1
			send(msg)
		}
		msg = rec()

		// leader was elected
		if msg.msg_type == BROADCAST {
			send(msg)
			output <- msg.id
			return
		}
	}

	// Leader broadcast
	msg.msg_type = BROADCAST
	send(msg)
	rec()

	output <- msg.id
}

func rec_template(in_id int, node lib.Node) func() Msg {
	return func() Msg {
		var decoded_msg Msg
		json.Unmarshal(node.ReceiveMessage(in_id), &decoded_msg)
		return decoded_msg
	}
}

func send_template(out_id int, node lib.Node) func(Msg) {
	return func(msg Msg) {
		encoded_msg, _ := json.Marshal(msg)
		node.SendMessage(out_id, encoded_msg)
	}
}

func run(node lib.Node) {
	left_to_right := make(chan int)
	right_to_left := make(chan int)
	id := node.GetIndex()

	go highamPrzytycka(
		id,
		rec_template(LEFT, node),
		send_template(RIGHT, node),
		left_to_right,
	)
	go highamPrzytycka(
		id,
		rec_template(RIGHT, node),
		send_template(LEFT, node),
		right_to_left,
	)

	leader := <-left_to_right
	_ = <-right_to_left

	setState(node, &state{
		leader: leader,
	})

	node.FinishProcessing(true)
}

func Run(nodes []lib.Node, runner lib.Runner) (int, int) {
	for _, node := range nodes {
		log.Println("Running node: ", node.GetIndex())
		go run(node)
	}

	runner.Run(true)
	check(nodes)
	return runner.GetStats()
}

func check(nodes []lib.Node) {
	maxId := nodes[0].GetIndex()
	theLeader := getState(nodes[0]).leader
	oneLeader := true

	for _, node := range nodes {
		maxId = max(maxId, node.GetIndex())
		if getState(node).leader != theLeader {
			oneLeader = false
		}
	}

	if theLeader == -1 {
		println("ERROR Invalid leader id!")
	}

	if !oneLeader {
		println("ERROR Multiple leaders elected!")
	}

	if maxId != theLeader {
		println("ERROR Leader doesn't have maximum id!")
	}
}

// helper functions
func getState(node lib.Node) state {
	var state state
	json.Unmarshal(node.GetState(), &state)
	return state
}

func setState(node lib.Node, state *state) {
	encodedState, _ := json.Marshal(*state)
	node.SetState(encodedState)
}
