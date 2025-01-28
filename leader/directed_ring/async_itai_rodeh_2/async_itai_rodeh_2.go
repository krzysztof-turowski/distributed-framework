package async_itai_rodeh_2

import (
	"encoding/binary"
	"log"
	"math"
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

func getDefaultRandomBit() IRandomBit {
	rng := rand.New(rand.NewSource(time.Now().UnixMilli()))
	return DefaultRandomBit{
		rand: rng,
	}
}

type MessageType byte

const (
	Elimination MessageType = iota
	Counter
	Elected
)

type message struct {
	MessageType MessageType
	Data        []byte
}

func (m message) Bytes() []byte {
	bytes := make([]byte, len(m.Data)+1)
	bytes[0] = byte(m.MessageType)
	copy(bytes[1:], m.Data)
	return bytes
}

func getMessage(bytes []byte) message {
	var msg message
	msg.MessageType = MessageType(bytes[0])
	msg.Data = make([]byte, len(bytes)-1)

	copy(msg.Data, bytes[1:])

	return msg
}

// Returns false if eliminated, true otherwise
func eliminationPhase(v lib.Node, c uint32, randomBit IRandomBit) bool {
	for i := uint32(0); i < c; i++ {
		bit := randomBit.ReadBit()

		v.SendMessage(0, message{
			MessageType: Elimination,
			Data:        []byte{bit},
		}.Bytes())

		msg := getMessage(v.ReceiveMessage(0))
		if msg.MessageType == Elimination && msg.Data[0] > bit {
			setState(v, NonLeader)
			return false
		}
	}
	return true
}

func verificationPhase(v lib.Node, n uint32) {
	bytes := binary.BigEndian.AppendUint32([]byte{}, 1)

	v.SendMessage(0, message{
		MessageType: Counter,
		Data:        bytes,
	}.Bytes())

	msg := getMessage(v.ReceiveMessage(0))

	if msg.MessageType == Counter && binary.BigEndian.Uint32(msg.Data) == n {
		setState(v, Leader)
		v.SendMessage(0, message{
			MessageType: Elected,
		}.Bytes())
	}
}

func inactive(v lib.Node) {
	for {
		bytes := v.ReceiveMessage(0)
		msg := getMessage(bytes)

		switch msg.MessageType {
		case Elimination:
			v.SendMessage(0, bytes)
		case Counter:
			rec := binary.BigEndian.Uint32(msg.Data)
			newBytes := binary.BigEndian.AppendUint32([]byte{}, rec+1)
			v.SendMessage(0, message{
				MessageType: Counter,
				Data:        newBytes,
			}.Bytes())
		case Elected:
			v.SendMessage(0, bytes)
			return
		}
	}

}

func run(v lib.Node, n uint32, c uint32, randomBit IRandomBit) {
	v.StartProcessing()
	initialize(v)
	log.Printf("%d initialized\n", v.GetIndex())
	for getState(v) == Unknown {
		log.Printf("%d started phase I\n", v.GetIndex())

		if eliminationPhase(v, c, randomBit) {
			verificationPhase(v, n)
		}
	}
	if getState(v) == NonLeader {
		inactive(v)
	}

	v.FinishProcessing(true)
}

func initialize(v lib.Node) {
	setState(v, Unknown)
}

func getState(v lib.Node) ModeType {
	return ModeType(v.GetState())
}

func setState(v lib.Node, state ModeType) {
	v.SetState([]byte(state))
}

func check(vertices []lib.Node) bool {
	var nonleaders int = 0

	for _, v := range vertices {
		state := getState(v)
		if state == NonLeader {
			nonleaders++
		}
	}

	leaders := len(vertices) - nonleaders
	if leaders < 1 {
		panic("Zero leaders were found")
	} else if leaders > 1 {
		panic("More than one active vertices. Maybe try again or change the number of stages constant")
	}

	return true
}

func startRunners(n int, c int, randomBit IRandomBit) (int, int) {
	vertices, runner := lib.BuildDirectedRing(n)
	for _, v := range vertices {
		go run(v, uint32(n), uint32(c), randomBit)
	}
	runner.Run(false)

	check(vertices)
	log.Println("Single leader elected")

	return runner.GetStats()
}

func Run(n int) (int, int) {
	c := int(math.Ceil(5 * math.Log2(float64(n))))
	return startRunners(n, c, getDefaultRandomBit())
}

type RandBit struct {
	rand *rand.Rand
	p    float64
}

func (r RandBit) ReadBit() uint8 {
	if rand.Float64() > r.p {
		return 0
	}
	return 1
}

func getRandBit(p float64) IRandomBit {
	return RandBit{
		rand.New(rand.NewSource(time.Now().UnixMilli())),
		p,
	}
}

func RunProb(n int, p float64) (int, int) {
	c := int(math.Ceil(5 * math.Log2(float64(n))))
	return startRunners(n, c, getRandBit(p))
}
