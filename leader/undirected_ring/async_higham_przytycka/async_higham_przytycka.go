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
	Msg_type int
	Id       int
	Round    int
	Counter  int
}

type state struct {
	Leader int
}

//

func nullMsg() Msg {
	return Msg{
		Msg_type: NORMAL,
		Id:       -1,
		Round:    -1,
		Counter:  0,
	}
}

func firstMsg(id int) Msg {
	return Msg{
		Msg_type: NORMAL,
		Id:       id,
		Round:    0,
		Counter:  0,
	}
}

func leaderTest(prev_msg Msg, msg Msg) bool {
	return msg.Id == prev_msg.Id && msg.Round == prev_msg.Round
}

func casualtyTest(prev_msg Msg, msg Msg) bool {
	return prev_msg.Round == msg.Round && ((odd(msg.Round) && msg.Id > prev_msg.Id) || (even(msg.Round) && msg.Id < prev_msg.Id))
}

func promotionTest(prev_msg Msg, msg Msg) bool {
	return (even(msg.Round) && prev_msg.Round == msg.Round-1 && msg.Id > prev_msg.Id) || (odd(msg.Round) && (msg.Counter == 0 || (prev_msg.Round == msg.Round && msg.Id < prev_msg.Id)))
}

func highamPrzytycka(proc_id int, rec func() Msg, send func(Msg), output chan<- int) {
	msg := firstMsg(proc_id)
	prev_msg := nullMsg()

	// Leader election
	for !leaderTest(prev_msg, msg) {
		if !casualtyTest(prev_msg, msg) {
			if promotionTest(prev_msg, msg) {
				msg.Round += 1
				msg.Counter = Fib(msg.Round + 2)
			}

			prev_msg = msg

			msg.Counter -= 1
			send(msg)
		}
		msg = rec()

		// check whether leader was elected
		if msg.Msg_type == BROADCAST {
			msg.Id = max(proc_id, msg.Id)
			send(msg)

			// broadcast max
			msg = rec()
			send(msg)
			output <- msg.Id
			return
		}
	}

	// Find max
	msg.Msg_type = BROADCAST
	msg.Id = max(proc_id, msg.Id)
	send(msg)
	msg = rec()

	// Broadcast max
	send(msg)
	rec()

	output <- msg.Id
}

func run(node lib.Node) {
	node.StartProcessing()

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
	<-right_to_left

	setState(node, state{
		leader,
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
	theLeader := getState(nodes[0]).Leader
	oneLeader := true

	for _, node := range nodes {
		maxId = max(maxId, node.GetIndex())
		if getState(node).Leader != theLeader {
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

	println("The leader: ", theLeader)
}

// helper functions
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

func getState(node lib.Node) state {
	var state state
	json.Unmarshal(node.GetState(), &state)
	return state
}

func setState(node lib.Node, state state) {
	encodedState, _ := json.Marshal(state)
	node.SetState(encodedState)
}
