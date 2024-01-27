package sync_metivier_c

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

// Run the Métivier et al. randomized MIS algorithm on a random graph
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
			sendMessage(node, i, bit)
		}
	}
}

func (s *state) receiveBits(node lib.Node) {
	for i, neighbor := range s.Neighbors {
		if !neighbor.Responsive {
			continue
		}
		received, _ := receiveMessage(node, i) // in this stage, bits are always sent, so it is safe to ignore the flag
		bit := s.getNeighborBit(neighbor)
		if bit != received { // symmetry is broken
			neighbor.Wins[neighbor.WinPhase] = !bit // we know whether we won this phase against this neighbor
			neighbor.WinPhase++                     // we can start the next phase of symmetry breaking
			neighbor.BitIndex = 0
		} else { // still symmetrical
			neighbor.BitIndex++ // we will have to send another bit
		}
	}
}

func (s *state) sendIns(node lib.Node) {
	in, known := s.canEnterMIS()
	if known {
		s.announce(node, in)
		s.InsSent = true
		s.Chosen = in
	} else {
		s.announceEmpty(node)
	}
}

func (s *state) receiveIns(node lib.Node) {
	for i, neighbor := range s.Neighbors {
		if !neighbor.Responsive {
			continue
		}
		if in, ok := receiveMessage(node, i); ok {
			neighbor.Ins[neighbor.InPhase] = in
			neighbor.InPhase++
		}
	}
}

func (s *state) sendOuts(node lib.Node) {
	out, known := s.shouldDisconnect()
	if known {
		s.announce(node, out)
		s.OutsSent = true
		s.Done = out
	} else {
		s.announceEmpty(node)
	}
}

func (s *state) receiveOuts(node lib.Node) {
	for i, neighbor := range s.Neighbors {
		if !neighbor.Responsive {
			continue
		}
		if out, ok := receiveMessage(node, i); ok {
			neighbor.Responsive = !out
			neighbor.Outs[neighbor.OutPhase] = out
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

// Returns (`enter`, `known`)
// If we have enough information to decide whether we should enter the MIS, `known` is `true` and `enter` represents the decision
// If we are missing some information, `known` is `false` and `enter` is undefined
func (s *state) canEnterMIS() (bool, bool) {
	if s.InsSent {
		return false, false
	}
	enter := true
	for _, neighbor := range s.Neighbors {
		win, ok := neighbor.Wins[s.Phase]
		if !ok {
			return false, false // we do not have enough information yet
		} else if !win {
			enter = false
		}
	}
	return enter, true
}

// Returns (`disconnect`, `known`)
// If we have enough information to decide whether we should disconnect, `known` is `true` and `disconnect` represents the decision
// If we are missing some information, `known` is `false` and `disconnect` is undefined
func (s *state) shouldDisconnect() (bool, bool) {
	if s.OutsSent {
		return false, false
	}
	disconnect := false
	for _, neighbor := range s.Neighbors {
		in, ok := neighbor.Ins[s.Phase]
		if !ok {
			return false, false // we do not have enough information yet
		} else if in {
			disconnect = true
		}
	}
	if s.Chosen {
		disconnect = true
	}
	return disconnect, true
}

func (s *state) advancePhase() {
	for _, neighbor := range s.Neighbors {
		if _, ok := neighbor.Outs[s.Phase]; !ok {
			return
		}
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
	s.InsSent, s.OutsSent = false, false
}

// Sends `bit` to all `Responsive` neighbors
func (s *state) announce(node lib.Node, bit bool) {
	for i, neighbor := range s.Neighbors {
		if neighbor.Responsive {
			sendMessage(node, i, bit)
		}
	}
}

// Sends an empty message to all `Responsive` neighbors
func (s *state) announceEmpty(node lib.Node) {
	for i, neighbor := range s.Neighbors {
		if neighbor.Responsive {
			sendEmptyMessage(node, i)
		}
	}
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

func sendMessage(node lib.Node, to int, bit bool) {
	var content byte = 0
	if bit {
		content = 1
	}
	node.SendMessage(to, []byte{content})
}

func sendEmptyMessage(node lib.Node, to int) {
	node.SendMessage(to, nil)
}

// Returns (`received`, `ok`)
// If there was a message, `ok` is `true` and `received` is set to the received bit
// If there was no message, `ok` is `false` and `received` is undefined
func receiveMessage(node lib.Node, from int) (bool, bool) {
	message := node.ReceiveMessage(from)
	if message == nil {
		return false, false
	}
	content := message[0] == 1
	return content, true
}
