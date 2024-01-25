package sync_metivier_et_al

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

func Run(n int, p float64) (int, int) {
	nodes, synchronizer := lib.BuildSynchronizedRandomGraph(n, p)
	for _, node := range nodes {
		log.Println("Node", node.GetIndex(), "about to run")
		go run(node)
	}
	synchronizer.Synchronize(0)
	check(nodes)
	return synchronizer.GetStats()
}

/* CHECKS */

func check(nodes []lib.Node) {
	for _, node := range nodes {
		chosen := getState(node).Chosen
		chosenNeighbor := none
		for _, neighbor := range node.GetOutNeighbors() {
			if getState(neighbor).Chosen {
				chosenNeighbor = neighbor.GetIndex()
			}
		}
		if chosen && chosenNeighbor != none {
			panic(fmt.Sprint("Node ", node.GetIndex(), " chosen alongside its neighbor ", chosenNeighbor))
		}
		if !chosen && chosenNeighbor == none {
			panic(fmt.Sprint("Node ", node.GetIndex(), " not chosen although available"))
		}
	}
}

/* THE SKELETON OF THE ALGORITHM */

func run(node lib.Node) {
	node.StartProcessing()
	finish := initialize(node)
	node.FinishProcessing(finish)

	for round := 1; !finish; round++ {
		node.StartProcessing()
		finish = process(node)
		node.FinishProcessing(finish)
	}
}

func initialize(node lib.Node) bool {
	setState(node, makeState(node.GetOutChannelsCount()))
	return false
}

func process(node lib.Node) bool {
	s := getState(node)
	switch s.Stage {
	case bitExchange:
		s.sendBits(node)
		s.receiveBits(node)
		s.Stage = inExchange
	case inExchange:
		s.sendIns(node)
		s.receiveIns(node)
		s.Stage = outExchange
	case outExchange:
		s.sendOuts(node)
		s.receiveOuts(node)
		s.advancePhase()
		s.Stage = bitExchange
	default:
		panic(fmt.Sprint("Invalid stage ", s.Stage))
	}
	done := s.Done
	setState(node, s)
	return done
}

type stage byte

const (
	bitExchange stage = iota
	inExchange
	outExchange
)

const none int = -1

/* OPERATIONS ON INTERNAL NODE STATES */

type state struct {
	Stage     stage
	Neighbors map[int]*neighbor
	Bits      map[int][]bool
	Phase     int
	InsSent   bool
	OutsSent  bool
	Chosen    bool
	Done      bool
}

func makeState(degree int) state {
	s := state{
		Stage:     bitExchange,
		Neighbors: make(map[int]*neighbor),
		Bits:      make(map[int][]bool),
		Phase:     0,
		InsSent:   false,
		OutsSent:  false,
		Chosen:    false,
		Done:      false,
	}
	for i := 0; i < degree; i++ {
		s.Neighbors[i] = makeNeighbor()
	}
	return s
}

type neighbor struct {
	WinPhase   int
	Wins       map[int]bool
	BitIndex   int
	InPhase    int
	Ins        map[int]bool
	OutPhase   int
	Outs       map[int]bool
	Responsive bool
}

func makeNeighbor() *neighbor {
	return &neighbor{
		WinPhase:   0,
		Wins:       make(map[int]bool),
		BitIndex:   0,
		InPhase:    0,
		Ins:        make(map[int]bool),
		OutPhase:   0,
		Outs:       make(map[int]bool),
		Responsive: true,
	}
}

func (s *state) sendBits(node lib.Node) {
	for i, neighbor := range s.Neighbors {
		if neighbor.Responsive {
			bit := s.getNeighborBit(neighbor)
			sendMessage(node, i, &bit)
		}
	}
}

func (s *state) receiveBits(node lib.Node) {
	for i, neighbor := range s.Neighbors {
		if !neighbor.Responsive {
			continue
		}
		received := *receiveMessage(node, i) // in this stage, bits are always sent, so it is safe to dereference
		bit := s.getNeighborBit(neighbor)
		if bit != received { // symmetry is broken
			neighbor.Wins[neighbor.WinPhase] = !bit // we know whether we won this phase against this neighbor
			neighbor.WinPhase++                    // we can start the next phase of symmetry breaking
			neighbor.BitIndex = 0
		} else { // still symmetrical
			neighbor.BitIndex++ // we will have to send another bit
		}
	}
}

func (s *state) sendIns(node lib.Node) {
	in := s.canEnterMIS()
	for i, neighbor := range s.Neighbors {
		if neighbor.Responsive {
			sendMessage(node, i, in)
		}
	}
	if in != nil {
		s.InsSent = true
		s.Chosen = *in
	}
}

func (s *state) receiveIns(node lib.Node) {
	for i, neighbor := range s.Neighbors {
		if !neighbor.Responsive {
			continue
		}
		in := receiveMessage(node, i)
		if in != nil {
			neighbor.Ins[neighbor.InPhase] = *in
			neighbor.InPhase++
		}
	}
}

func (s *state) sendOuts(node lib.Node) {
	out := s.shouldDisconnect()
	for i, neighbor := range s.Neighbors {
		if neighbor.Responsive {
			sendMessage(node, i, out)
		}
	}
	if out != nil {
		s.OutsSent = true
		s.Done = *out
	}
}

func (s *state) receiveOuts(node lib.Node) {
	for i, neighbor := range s.Neighbors {
		if !neighbor.Responsive {
			continue
		}
		out := receiveMessage(node, i)
		if out != nil {
			neighbor.Responsive = !*out
			neighbor.Outs[neighbor.OutPhase] = *out
			neighbor.OutPhase++
		}
	}
}

func (s *state) getNeighborBit(neighbor *neighbor) bool {
	return s.getBit(neighbor.WinPhase, neighbor.BitIndex)
}

func (s *state) getBit(phase int, index int) bool {
	bits := s.Bits[phase]
	if index > len(bits) {
		panic("Earlier bits have not been drawn yet")
	}
	if index == len(bits) {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		bits = append(bits, rng.Intn(2) == 1)
	}
	s.Bits[phase] = bits
	return bits[index]
}

func (s *state) canEnterMIS() *bool {
	if s.InsSent {
		return nil
	}
	all := true
	enter := true
	for _, neighbor := range s.Neighbors {
		win, ok := neighbor.Wins[s.Phase]
		if !ok {
			all = false
		} else if !win {
			enter = false
		}
	}
	if !all {
		return nil
	}
	return &enter
}

func (s *state) shouldDisconnect() *bool {
	if s.OutsSent {
		return nil
	}
	all := true
	disconnect := false
	for _, neighbor := range s.Neighbors {
		in, ok := neighbor.Ins[s.Phase]
		if !ok {
			all = false
		} else if in {
			disconnect = true
		}
	}
	if !all {
		return nil
	}
	if s.Chosen {
		disconnect = true
	}
	return &disconnect
}

func (s *state) advancePhase() {
	all := true
	for _, neighbor := range s.Neighbors {
		_, ok := neighbor.Outs[s.Phase]
		if !ok {
			all = false
		}
	}
	if !all {
		return
	}
	for i, neighbor := range s.Neighbors {
		if neighbor.Outs[s.Phase] {
			delete(s.Neighbors, i)
		} else {
			delete(neighbor.Wins, s.Phase)
			delete(neighbor.Ins, s.Phase)
			delete(neighbor.Outs, s.Phase)
		}
	}
	delete(s.Bits, s.Phase)
	s.Phase++
	s.InsSent = false
	s.OutsSent = false
}

/* GENERAL UTILS */

func setState(node lib.Node, s state) {
	representation, _ := json.Marshal(s)
	node.SetState(representation)
}

func getState(node lib.Node) state {
	var s state
	json.Unmarshal(node.GetState(), &s)
	return s
}

func sendMessage(node lib.Node, to int, bit *bool) {
	var message []byte = nil
	if bit != nil {
		var content byte = 0
		if *bit {
			content = 1
		}
		message = []byte{content}
	}
	node.SendMessage(to, message)
}

func receiveMessage(node lib.Node, from int) *bool {
	message := node.ReceiveMessage(from)
	if message == nil {
		return nil
	}
	content := message[0] == 1
	return &content
}
