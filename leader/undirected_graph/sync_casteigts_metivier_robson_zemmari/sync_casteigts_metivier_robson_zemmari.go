package sync_casteigts_metivier_robson_zemmari

import (
	"encoding/json"
	"fmt"
	"log"
	"math/bits"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

// Run the Casteigts-Métivier-Robson-Zemmari deterministic leader election algorithm on a random network
func Run(n int, p float64) (int, int) {
	possibleEdges := (n * (n - 1)) / 2
	m := int(float64(possibleEdges) * p)
	m = max(m, n-1)
	m = min(m, possibleEdges)

	nodes, synchronizer := lib.BuildSynchronizedRandomConnectedGraphWithUniqueIndices(n, m)
	for _, node := range nodes {
		log.Println("Node", node.GetIndex(), "about to run")
		go run(node) // Every node is assigned a unique identifier of O(log(n)) bits
	}
	synchronizer.Synchronize(0)
	check(nodes)
	return synchronizer.GetStats()
}

/* CHECKS */

func check(nodes []lib.Node) {
	maxIndex := 0
	leader := none
	for _, node := range nodes {
		s := getState(node)
		index := node.GetIndex()
		maxIndex = max(maxIndex, index)
		if !s.Elected {
			continue
		}
		if leader != none {
			panic("Multiple leaders")
		}
		leader = index
	}
	if leader == none {
		panic("No leader")
	}
	if leader != maxIndex {
		panic("The leader does not have the greatest index")
	}
	maxAlpha := toAlphaEncoding(maxIndex)
	for _, node := range nodes {
		if !equal(getState(node).Prefix, maxAlpha) {
			panic("A node does not know the leader's identifier")
		}
	}
}

/* THE SKELETON OF THE ALGORITHM */

func run(node lib.Node) {
	node.StartProcessing()
	s := newState(node.GetIndex(), node.GetOutChannelsCount())
	setState(node, s)
	node.FinishProcessing(false)

	finish := false
	for round := 1; !finish; round++ {
		node.StartProcessing()
		finish = process(node)
		node.FinishProcessing(finish)
	}
}

func process(node lib.Node) bool {
	s := getState(node)

	switch s.Phase {
	case modificationExchange:
		s.ExchangeRound++
		operation := s.updatePrefixAndParent()
		announceModification(node, operation, &s)
		if operation == shutdown {
			node.IgnoreFutureMessages()
			return true
		}
		processIncomingModifications(node, &s)
	case externalTerminationUpdate:
		announceTermination(node, &s)
		processIncomingTerminations(node, &s)
	case internalTerminationUpdate:
		candidate := s.isCandidate()
		if !s.Done && s.shouldSetTermination() {
			s.Termination = true
		}
		reportTerminationToParent(node, &s)
		collectTerminationsFromChildren(node, &s, candidate)
	}

	s.Phase = nextPhase(s.Phase)

	setState(node, s)

	return false
}

func announceModification(node lib.Node, operation operation, s *state) {
	for i := range s.Neighbors {
		neighbor := &s.Neighbors[i]
		if neighbor.Done {
			continue
		}
		sendModification(node, i, modification{operation, i == s.Parent})
		if operation == shutdown {
			neighbor.Done = true
		}
	}
}

func processIncomingModifications(node lib.Node, s *state) {
	for i, neighbor := range s.Neighbors {
		if neighbor.Done {
			continue
		}
		s.processModification(i, receiveModification(node, i))
	}
}

func announceTermination(node lib.Node, s *state) {
	for i, neighbor := range s.Neighbors {
		if neighbor.Done {
			continue
		}
		sendTermination(node, i, s.Termination)
	}
}

func processIncomingTerminations(node lib.Node, s *state) {
	for i, neighbor := range s.Neighbors {
		if neighbor.Done {
			continue
		}
		s.Neighbors[i].Termination = receiveTermination(node, i)
	}
}

func reportTerminationToParent(node lib.Node, s *state) {
	if s.Parent != none {
		sendTermination(node, s.Parent, s.Termination)
	}
}

func collectTerminationsFromChildren(node lib.Node, s *state, candidate bool) {
	persistentTerm := 0
	for i, neighbor := range s.Neighbors {
		if !neighbor.IsChild {
			continue
		}
		term := receiveTermination(node, i)
		if neighbor.Termination && term {
			persistentTerm++
		}
		neighbor.Termination = term
	}
	if candidate && persistentTerm == len(s.Neighbors) {
		log.Println(s.PlainID, "becomes leader")
		s.Elected = true
		s.Done = true
	}
}

type phase byte

const (
	modificationExchange phase = iota
	externalTerminationUpdate
	internalTerminationUpdate
)

func nextPhase(phase phase) phase {
	switch phase {
	case modificationExchange:
		return externalTerminationUpdate
	case externalTerminationUpdate:
		return internalTerminationUpdate
	case internalTerminationUpdate:
		return modificationExchange
	default:
		panic(fmt.Sprint("Not a phase: ", phase))
	}
}

/* OPERATIONS ON INTERNAL NODE STATES */

type bit byte // This looks funny, but it is much more convenient than operating on actual bits

type operation byte

const (
	append0 operation = iota
	append1
	delete1
	delete2
	delete3
	change
	null
	shutdown
)

const none int = -1

type state struct {
	ExchangeRound int
	Phase         phase
	PlainID       int
	Active        bool  // Is it still possible for this node to become the leader?
	AlphaID       []bit // This node's own identifier (alpha-encoded)
	Prefix        []bit // The prefix of the leader's identifier that this node knows
	Neighbors     []neighbor
	Parent        int // Index of the parent among the neighbors
	Elected       bool
	Termination   bool
	Done          bool
}

// Note that these messages have (semantically) constant size, i.e. we do not send identifiers

type modification struct {
	Operation operation // The operation that the sender performed on its known prefix
	FromChild bool      // Does the sender want the receiver to be its parent in the spanning tree?
}

type termination struct {
	Flag bool
}

// What the node knows about its neighbors
type neighbor struct {
	Prefix        []bit // The last known prefix from this neighbor's state
	IsChild       bool
	LastOperation operation
	Termination   bool
	Done          bool
}

func newNeighbor() neighbor {
	return neighbor{
		Prefix:        make([]bit, 0),
		IsChild:       false,
		LastOperation: null,
		Termination:   false,
		Done:          false,
	}
}

func newState(id int, degree int) state {
	neighbors := make([]neighbor, degree)
	for i := range neighbors {
		neighbors[i] = newNeighbor()
	}
	return state{
		ExchangeRound: 0,
		Phase:         modificationExchange,
		PlainID:       id,
		Active:        true,
		AlphaID:       toAlphaEncoding(id),
		Prefix:        make([]bit, 0),
		Neighbors:     neighbors,
		Parent:        none,
		Elected:       false,
		Termination:   false,
		Done:          false,
	}
}

func (s *state) updatePrefixAndParent() operation {
	if s.Done {
		return shutdown
	}
	operation := null
	if toDelete := s.possibleDelete(); toDelete != 0 {
		log.Println("round", s.ExchangeRound, ":", s.PlainID, "uses rule 1: delete", toDelete)
		operation = deleteOperation(toDelete)
	} else if parent := s.possibleChange(); parent != none {
		log.Println("round", s.ExchangeRound, ":", s.PlainID, "uses rule 2: change")
		s.Active = false
		operation = change
		s.Parent = parent
	} else if parent := s.possibleAppend(1); parent != none {
		log.Println("round", s.ExchangeRound, ":", s.PlainID, "uses rule 3: append 1")
		operation = append1
		s.Parent = parent
	} else if parent := s.possibleAppend(0); parent != none {
		log.Println("round", s.ExchangeRound, ":", s.PlainID, "uses rule 4: append 0 ")
		operation = append0
		s.Parent = parent
	} else if toAppend, ok := s.possibleExtend(); ok {
		log.Println("round", s.ExchangeRound, ":", s.PlainID, "uses rule 5: append", toAppend)
		operation = appendOperation(toAppend)
	}
	s.Prefix = apply(s.Prefix, operation)
	if operation != null {
		s.Termination = false
	}
	return operation
}

func (s *state) processModification(from int, message modification) {
	neighbor := &s.Neighbors[from]
	if message.Operation == shutdown {
		s.Done = true
		neighbor.Done = true
		return
	}
	neighbor.Prefix = apply(neighbor.Prefix, message.Operation)
	neighbor.LastOperation = message.Operation
	if message.Operation != null {
		neighbor.Termination = false
	}
	if message.FromChild {
		if !neighbor.IsChild {
			s.Termination = false
		}
		neighbor.IsChild = true
	} else {
		neighbor.IsChild = false
	}
}

/* RULES FOR STATE MANIPULATION */

func (s *state) possibleDelete() int {
	return max(s.possibleDeleteA(), s.possibleDeleteB())
}

func (s *state) possibleDeleteA() int {
	toDelete := 0
	for _, neighbor := range s.Neighbors {
		if isProperPrefixOf(neighbor.Prefix, s.Prefix) && isDeleteOperation(neighbor.LastOperation) {
			toDelete = max(toDelete, min(len(s.Prefix)-len(neighbor.Prefix), 3))
		}
	}
	return toDelete
}

func (s *state) possibleDeleteB() int {
	toDelete := 0
	for _, neighbor := range s.Neighbors {
		common := longestCommonPrefix(s.Prefix, neighbor.Prefix)
		diff := len(common)
		validPosition := diff < len(s.Prefix)-1 && diff < len(neighbor.Prefix)
		if validPosition && s.Prefix[diff] == 0 && neighbor.Prefix[diff] == 1 {
			toDelete = max(toDelete, len(s.Prefix)-1-diff)
		}
	}
	return toDelete
}

func (s *state) possibleChange() int {
	if len(s.Prefix) == 0 {
		return none
	}
	last := len(s.Prefix) - 1
	if s.Prefix[last] != 0 {
		return none
	}
	init := s.Prefix[:last]
	for i, neighbor := range s.Neighbors {
		if isProperPrefixOf(init, neighbor.Prefix) && neighbor.Prefix[last] == 1 {
			return i
		}
	}
	return none
}

func (s *state) possibleAppend(bit bit) int {
	for i, neighbor := range s.Neighbors {
		if isProperPrefixOf(s.Prefix, neighbor.Prefix) && neighbor.Prefix[len(s.Prefix)] == bit {
			return i
		}
	}
	return none
}

func (s *state) possibleExtend() (bit, bool) {
	if s.Active && s.ExchangeRound <= len(s.AlphaID) {
		return s.AlphaID[s.ExchangeRound-1], true
	}
	return 0, false
}

func (s *state) shouldSetTermination() bool {
	if s.Active || s.Termination || !isWellFormed(s.Prefix) {
		return false
	}
	for _, neighbor := range s.Neighbors {
		if !equal(neighbor.Prefix, s.Prefix) {
			return false
		}
		if neighbor.IsChild && !neighbor.Termination {
			return false
		}
	}
	return true
}

func (s *state) isCandidate() bool {
	for _, neighbor := range s.Neighbors {
		if !equal(neighbor.Prefix, s.Prefix) || !neighbor.Termination {
			return false
		}
	}
	return isWellFormed(s.Prefix)
}

/* OPERATIONS ON BIT STRINGS */

func equal(a, b []bit) bool {
	return len(a) == len(b) && isPrefixOf(a, b)
}

func isProperPrefixOf(a, b []bit) bool {
	return len(a) < len(b) && isPrefixOf(a, b)
}

func isPrefixOf(a, b []bit) bool {
	return len(longestCommonPrefix(a, b)) == len(a)
}

func longestCommonPrefix(a, b []bit) []bit {
	commonLength := min(len(a), len(b))
	i := 0
	for ; i < commonLength; i++ {
		if a[i] != b[i] {
			break
		}
	}
	return a[:i]
}

func isWellFormed(candidate []bit) bool {
	for i, bit := range candidate {
		if bit == 0 {
			return len(candidate) == 2*i+1
		}
	}
	return false
}

// number -> binary -> [1^(len binary)][0][binary]
// e.g. 5 -> 101 -> 1110101
func toAlphaEncoding(id int) []bit {
	var msb int = 0
	if id != 0 {
		msb = bits.UintSize - 1 - bits.LeadingZeros(uint(id))
	}
	encoded := make([]bit, 0, 2*msb+1)
	for i := 0; i <= msb; i++ {
		encoded = append(encoded, 1)
	}
	encoded = append(encoded, 0)
	for i := msb; i >= 0; i-- {
		var bit bit = 0
		if (id & (1 << i)) != 0 {
			bit = 1
		}
		encoded = append(encoded, bit)
	}
	return encoded
}

// The following four functions could have been written in a shorter way, but this is more explicit
// Moreover, it will still work even if we decide to reorder the constants

func isDeleteOperation(operation operation) bool {
	return operation == delete1 || operation == delete2 || operation == delete3
}

func deleteOperation(n int) operation {
	switch n {
	case 1:
		return delete1
	case 2:
		return delete2
	case 3:
		return delete3
	default:
		panic(fmt.Sprint("Cannot delete ", n, " bits: no such operation"))
	}
}

func appendOperation(bit bit) operation {
	switch bit {
	case 0:
		return append0
	case 1:
		return append1
	default:
		panic(fmt.Sprint("Cannot append ", bit, ": not a valid bit"))
	}
}

func apply(sequence []bit, operation operation) []bit {
	end := len(sequence)
	switch operation {
	case append0:
		return append(sequence, 0)
	case append1:
		return append(sequence, 1)
	case delete1:
		return sequence[:end-1]
	case delete2:
		return sequence[:end-2]
	case delete3:
		return sequence[:end-3]
	case change:
		sequence[end-1] = 1
	}
	return sequence
}

/* GENERAL UTILS */

func sendModification(node lib.Node, to int, message modification) {
	representation, _ := json.Marshal(message)
	node.SendMessage(to, representation)
}

func receiveModification(node lib.Node, from int) modification {
	var message modification
	json.Unmarshal(node.ReceiveMessage(from), &message)
	return message
}

func sendTermination(node lib.Node, to int, flag bool) {
	message := termination{flag}
	representation, _ := json.Marshal(message)
	node.SendMessage(to, representation)
}

func receiveTermination(node lib.Node, from int) bool {
	var message termination
	json.Unmarshal(node.ReceiveMessage(from), &message)
	return message.Flag
}

func setState(node lib.Node, s state) {
	representation, _ := json.Marshal(s)
	node.SetState(representation)
}

func getState(node lib.Node) state {
	var s state
	json.Unmarshal(node.GetState(), &s)
	return s
}
