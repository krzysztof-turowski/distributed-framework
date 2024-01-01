package async_itai_rodeh

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"log"
	"math"
	"math/rand"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

const (
	LEFT  = 0
	RIGHT = 1
)

type state struct {
	Counter          int
	Is_active        bool
	Is_leader        bool
	Is_leader_chosen bool
	Bit              bool
}

type messageData struct {
	Msg_type  uint8
	Bit       bool
	Hop_count uint64
}

//Ring start and answer check

func Run(nodes []lib.Node, runner lib.Runner) (int, int) {
	for _, node := range nodes {
		log.Println("Running node", node.GetIndex())
		go run(node)
	}
	runner.Run(true)

	leaders := 0

	for _, node := range nodes {
		if getState(node).Is_leader {
			leaders += 1
		}
	}
	if leaders == 0 {
		panic("No leader elected")
	}
	if leaders > 1 {
		panic("More than one leader elected")
	}

	log.Println("---OK--- exactly one leader elected")

	return runner.GetStats()
}

// Main function run in each goroutine

func run(node lib.Node) {
	node.StartProcessing()
	initialize(node)
	if process(node) {
		log.Println("Node", node.GetIndex(), "is elected a leader")
	}
	node.FinishProcessing(true)
}

// Main algorithm loop

func process(node lib.Node) bool {
	for {
		var msg messageData
		sender, byteMsg := node.ReceiveAnyMessage()
		decodeAll(byteMsg, &msg)

		switch msg.Msg_type {
		case 0:
			handleItaiRodeh(node, sender, msg)
		case 1:
			handleVerificationMessage(node, sender, msg)
		case 2:
			handleLeaderProclamation(node, sender, msg)
		}

		if getState(node).Is_leader_chosen {
			return getState(node).Is_leader
		}
	}
}

func initialize(node lib.Node) {

	setState(node, &state{
		Counter:          0,
		Is_active:        true,
		Is_leader:        false,
		Is_leader_chosen: false,
		Bit:              (rand.Intn(2) == 1),
	})

	msg := messageData{
		Msg_type:  0,
		Bit:       getState(node).Bit,
		Hop_count: 0,
	}

	node.SendMessage(LEFT, encodeAll(msg))
	node.SendMessage(RIGHT, encodeAll(msg))
}

func handleItaiRodeh(node lib.Node, sender int, msg messageData) {

	S := getState(node)
	if S.Is_active { //active
		var lbit bool
		var rbit bool

		//waiting for second msg
		if sender == LEFT {
			lbit = msg.Bit
			decodeAll(node.ReceiveMessage(RIGHT), &msg)
			rbit = msg.Bit
		} else {
			rbit = msg.Bit
			decodeAll(node.ReceiveMessage(LEFT), &msg)
			lbit = msg.Bit
		}
		if S.Counter < 5*int(math.Log2(float64(node.GetSize()))) { //still normal round
			if !S.Bit && (lbit || rbit) { //we are being killed
				S.Is_active = false
			} else { //nobody dies
				S.Bit = (rand.Intn(2) == 1)
				S.Counter += 1
				msg.Bit = S.Bit
				node.SendMessage(LEFT, encodeAll(msg))
				node.SendMessage(RIGHT, encodeAll(msg))
			}
		} else { //last round, send verification msg
			msg.Msg_type = 1
			node.SendMessage(LEFT, encodeAll(msg))
			node.SendMessage(RIGHT, encodeAll(msg))
		}
	} else { //passive
		node.SendMessage((sender+1)%2, encodeAll(msg))
	}
	setState(node, &S)
}

func handleVerificationMessage(node lib.Node, sender int, msg messageData) {
	S := getState(node)
	if S.Is_active { //active
		var lhop uint64
		var rhop uint64

		//waiting for second msg
		if sender == LEFT {
			lhop = msg.Hop_count
			decodeAll(node.ReceiveMessage(RIGHT), &msg)
			rhop = msg.Hop_count
		} else {
			rhop = msg.Hop_count
			decodeAll(node.ReceiveMessage(LEFT), &msg)
			lhop = msg.Hop_count
		}

		if lhop == uint64(node.GetSize()-1) && rhop == uint64(node.GetSize()-1) { //our msg returned, so we are the leader
			S.Is_leader = true
			msg.Msg_type = 2 //leader proclamation msg
			node.SendMessage(RIGHT, encodeAll(msg))
		} else { //there is more than one leader - we reset the counter and initialize another sequence
			S.Counter = 0
			S.Bit = (rand.Intn(2) == 1)
			msg.Msg_type = 0
			msg.Bit = S.Bit
			node.SendMessage(LEFT, encodeAll(msg))
			node.SendMessage(RIGHT, encodeAll(msg))
		}
	} else { //passive
		msg.Hop_count += 1
		node.SendMessage((sender+1)%2, encodeAll(msg))
	}
	setState(node, &S)
}

func handleLeaderProclamation(node lib.Node, sender int, msg messageData) {
	S := getState(node)
	S.Is_leader_chosen = true

	if !S.Is_leader {
		node.SendMessage((sender+1)%2, encodeAll(msg))
	}
	setState(node, &S)
}

//----------------- helper functions -----------------

//state handling

func getState(node lib.Node) state {
	var state state
	json.Unmarshal(node.GetState(), &state)
	return state
}

func setState(node lib.Node, state *state) {
	encodedState, _ := json.Marshal(*state)
	node.SetState(encodedState)
}

//encoding and decoding

func encodeAll(values ...interface{}) []byte {
	buffer := bytes.Buffer{}
	encode(&buffer, values...)
	return buffer.Bytes()
}

func decodeAll(data []byte, values ...interface{}) {
	buffer := bytes.NewBuffer(data)
	decode(buffer, values...)
}

func encode(buffer *bytes.Buffer, values ...interface{}) {
	for _, value := range values {
		if binary.Write(buffer, binary.BigEndian, value) != nil {
			log.Println("DUUUPA   ", value)
			panic("Failed to encode a value")
		}
	}
}

func decode(buffer *bytes.Buffer, values ...interface{}) {
	for _, value := range values {
		if binary.Read(buffer, binary.BigEndian, value) != nil {
			panic("Failed to decode a value")
		}
	}
}
