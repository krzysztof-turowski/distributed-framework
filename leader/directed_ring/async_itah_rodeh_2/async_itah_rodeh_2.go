package async_itah_rodeh_2

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	. "github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/common"
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

type IRandomBit interface {
	ReadBit() uint8
}

type DefaultRandomBit struct {
	rand *rand.Rand
}

func (randomBit DefaultRandomBit) ReadBit() uint8 {
	return uint8(randomBit.rand.Intn(2))
}

func GetDefaultRandomBit() IRandomBit {
	rng := rand.New(rand.NewSource(time.Now().UnixMilli()))
	return DefaultRandomBit{
		rand: rng,
	}
}

type State struct {
	Status ModeType
	Leader int
}

type MessageType uint8

const (
	Bit = iota
	Check
	Elected
)

type message struct {
	MessageType MessageType
	Number      uint32
	Leader      int
}

func (m message) Bytes() []byte {
	bytes, _ := json.Marshal(m)
	return bytes
}

func getMessage(bytes []byte) message {
	var msg message
	json.Unmarshal(bytes, &msg)
	return msg
}

func eliminationPhase(v lib.Node, randomBit IRandomBit) {
	bit := randomBit.ReadBit()

	v.SendMessage(0, message{
		MessageType: Bit,
		Number:      uint32(bit),
	}.Bytes())

	msg := getMessage(v.ReceiveMessage(0))
	if msg.MessageType == Bit && msg.Number > uint32(bit) {
		setState(v, State{Status: Relay})
		log.Println("RETIRED")
	} else {
		setState(v, State{Status: Pass})
	}
}

func verificationPhase(v lib.Node, n uint32) {
	v.SendMessage(0, message{
		MessageType: Check,
		Number:      1,
	}.Bytes())

	msg := getMessage(v.ReceiveMessage(0))

	if msg.MessageType == Check && msg.Number == n {
		setState(v, State{
			Status: Leader,
			Leader: v.GetIndex(),
		})
		v.SendMessage(0, message{
			MessageType: Elected,
			Leader:      v.GetIndex(),
		}.Bytes())
	} else {
		setState(v, State{Status: Unknown})
	}
}

func idle(v lib.Node) {
	bytes := v.ReceiveMessage(0)
	msg := getMessage(bytes)

	switch msg.MessageType {
	case Bit:
		v.SendMessage(0, bytes)
	case Check:
		v.SendMessage(0, message{
			MessageType: Check,
			Number:      msg.Number + 1,
		}.Bytes())
	case Elected:
		v.SendMessage(0, bytes)
		setState(v, State{
			Status: NonLeader,
			Leader: msg.Leader,
		})
	}
}

func run(v lib.Node, n uint32, randomBit IRandomBit) {
	v.StartProcessing()
	initialize(v)
	log.Printf("%d initialized\n", v.GetIndex())
	for {
		switch status := getState(v).Status; status {
		case Unknown:
			eliminationPhase(v, randomBit)
		case Pass:
			verificationPhase(v, n)
		case Relay:
			idle(v)
		default:
			log.Printf("%d finishing %s\n", v.GetIndex(), status)
			goto end
		}
	}
end:
	v.FinishProcessing(true)
}

func initialize(v lib.Node) {
	setState(v, State{
		Status: Unknown,
	})
}

func getState(v lib.Node) State {
	var state State
	json.Unmarshal(v.GetState(), &state)
	return state
}

func setState(v lib.Node, state State) {
	data, _ := json.Marshal(state)
	v.SetState(data)
}

func check(vertices []lib.Node) {
	var leader int = 0

	for _, v := range vertices {
		state := getState(v)
		if state.Leader == 0 {
			panic("Not all aware of leader")

		} else if leader == 0 {
			leader = state.Leader
		}

		if leader != state.Leader {
			panic("Not all the same leaders")
		}
	}
}

func Run(n int, randomBit IRandomBit) (int, int) {
	vertices, runner := lib.BuildDirectedRing(n)

	for _, v := range vertices {
		go run(v, uint32(n), randomBit)
	}

	runner.Run(false)
	check(vertices)
	log.Println("Single leader elected")

	return runner.GetStats()
}

func RunDefault(n int) (int, int) {
	return Run(n, GetDefaultRandomBit())
}
