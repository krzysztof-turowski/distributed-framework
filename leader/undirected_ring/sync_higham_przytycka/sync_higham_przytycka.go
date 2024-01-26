package sync_higham_przytycka

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

const DEBUG_ID_MOD = 1000
const FIB64_UPPER_LIM = 93
const LOG_PREFIX = "[%d:%01d] "

type ID = int
type direction = int
type step = uint64

const (
	Left  direction = iota
	Right direction = iota
)

// region message types
type uhpEnvelope struct {
	Id    ID
	Dir   direction
	Round uint
	Cnt   step
	Max   ID
}
type uhpBroadcast = ID

type uhpMessage struct {
	ENVELOPE  *uhpEnvelope
	BROADCAST *uhpBroadcast
}

// endregion

// region state struct
type undirectedHighamPrzytycka struct {
	IsLeader   bool // dependent on both processors
	Processors [2]uhpProcessor
}

// used as namespace
type uhp = undirectedHighamPrzytycka
type uhpState = uhp

// endregion

// region processor

type status = byte

const (
	Active     status = iota
	TempLeader status = iota
	Dead       status = iota
	TrueLeader status = iota
)

// region Promoter
type uhpPromotionStrategy interface {
	Test(last uhpEnvelope, curr uhpEnvelope) (drop bool, promote bool)
	Promote(newEnvelope *uhpEnvelope)
}

// region Promoter::simple
type simplePromoter struct{}

func (simplePromoter) Test(last uhpEnvelope, curr uhpEnvelope) (drop bool, promote bool) {
	promote = curr.Round == last.Round
	drop = promote &&
		(curr.Round%2 == 0 && curr.Id > last.Id) ||
		(curr.Round%2 == 1 && curr.Id < last.Id)
	return drop, promote
}

func (simplePromoter) Promote(newEnvelope *uhpEnvelope) {
	newEnvelope.Round += 1
}

// endregion

// region Promoter::hp
type hpPromoter struct {
	F [FIB64_UPPER_LIM]step
}

func (hpPromoter) Test(last uhpEnvelope, curr uhpEnvelope) (drop bool, promote bool) {
	drop = curr.Round == last.Round // equal rounds are a necessary condition to drop envelope
	switch curr.Round % 2 {
	case 0: // even
		drop = drop && curr.Id < last.Id
		promote = last.Round == curr.Round-1
	case 1: // odd
		promote = drop || curr.Cnt == 0
		drop = drop && curr.Id > last.Id
	}
	promote = promote && !drop // promote implies !drop
	return drop, promote
}

func (p *hpPromoter) fib(n int) step {
	if n <= 1 {
		return step(n)
	}
	if p.F[n] != 0 {
		return p.F[n]
	}
	return p.fib(n-1) + p.fib(n-2)
}

func (p *hpPromoter) Promote(newEnvelope *uhpEnvelope) {
	newEnvelope.Round += 1
	newEnvelope.Cnt = p.fib(int(newEnvelope.Round + 2))
}

// endregion
// endregion

type uhpProcessor struct {
	Id       ID
	Dir      direction // for observability
	Last     uhpEnvelope
	Status   status
	Promoter uhpPromotionStrategy
}

// endregion

// region uhp:: (scope: bidirectional)

// region uhp::state
func (uhp) getState(node lib.Node) (state uhpState) {
	json.Unmarshal(node.GetState(), &state)
	// couldn't figure out how to serialize interface nicely
	state.Processors[Left].Promoter = &hpPromoter{}
	state.Processors[Right].Promoter = state.Processors[Left].Promoter
	return state
}

func (uhp) updateState(node lib.Node, state *uhpState) {
	payload, _ := json.Marshal(state)
	node.SetState(payload)
}

// endregion

func (uhp) send(d direction, node lib.Node, state *uhpState, msg *uhpMessage) {
	if msg == nil {
		return
	}
	if (*msg == uhpMessage{}) {
		node.SendMessage(d, nil)
	} else {
		payload, _ := json.Marshal(msg)
		node.SendMessage(d, payload)
	}
}

func (uhp) receive(d direction, node lib.Node, state *uhpState) *uhpMessage {
	var msg *uhpMessage
	json.Unmarshal(node.ReceiveMessage(d), &msg)
	if msg == nil {
		return &uhpMessage{}
	}
	return msg
}

// region uhp::constructors
func (uhp) newInstance(node lib.Node) *uhpState {
	id := node.GetIndex()
	res := &uhpState{
		IsLeader: false,
		Processors: [2]uhpProcessor{
			uhp{}.newProcessor(id, Left),
			uhp{}.newProcessor(id, Right),
		},
	}
	return res
}

func (uhp) newEnvelope(id int, d direction) uhpEnvelope {
	// id = id % DEBUG_ID_MOD
	// log.Printf(LOG_PREFIX+"created envelope", id, d)
	return uhpEnvelope{
		Id:    id,
		Dir:   d,
		Round: 0,
		Max:   id,
	}
}

func (uhp) newProcessor(id int, d direction) uhpProcessor {
	log.Printf(LOG_PREFIX+"created processor", id, d)
	return uhpProcessor{
		Id:       id,
		Dir:      d,
		Status:   Active,
		Promoter: &hpPromoter{},
	}
}

// endregion

func (uhp) initiateElection(node lib.Node) {
	state := uhp{}.newInstance(node)

	singlePass := func(p *uhpProcessor) {
		p.Last = uhp{}.newEnvelope(p.Id, p.Dir)
		uhp{}.send(p.Dir, node, state, &uhpMessage{ENVELOPE: &p.Last})
	}

	singlePass(&state.Processors[Left])
	singlePass(&state.Processors[Right])
	uhp{}.updateState(node, state)
}

func (uhp) round(node lib.Node) (finished bool) {
	state := uhp{}.getState(node)
	defer uhp{}.updateState(node, &state)

	if state.Processors[Left].testFinished() &&
		state.Processors[Right].testFinished() {
		state.IsLeader = state.Processors[Left].Status == TrueLeader &&
			state.Processors[Right].Status == TrueLeader
		return true
	}

	// log.Printf("[%d] State before round: %#v", node.GetIndex(), state)

	mailbox := [2]*uhpMessage{nil, nil}
	if state.Processors[Left].testReceive() {
		mailbox[Left] = uhp{}.receive(Right, node, &state)
	}
	if state.Processors[Right].testReceive() {
		mailbox[Right] = uhp{}.receive(Left, node, &state)
	}
	mailbox[Left] = state.Processors[Left].process(mailbox[Left])
	mailbox[Right] = state.Processors[Right].process(mailbox[Right])
	uhp{}.send(Left, node, &state, mailbox[Left])
	uhp{}.send(Right, node, &state, mailbox[Right])

	// log.Printf("[%d] State after round: %#v", node.GetIndex(), state)

	return false
}

// endregion

// region uhp::Processor (scope: unidirectional)
// region uhp::Processor::test
func (proc uhpProcessor) testLeader(curr uhpEnvelope) bool {
	// envelope returned to owner
	return curr == proc.Last
}

func (proc uhpProcessor) testReceive() bool {
	switch proc.Status {
	case Active:
		return true
	case TempLeader:
		return true
	default:
		return false
	}
}

func (proc uhpProcessor) testFinished() bool {
	switch proc.Status {
	case Active:
		return false
	case TempLeader:
		return false
	case TrueLeader:
		return true
	case Dead:
		return true
	default:
		return true
	}
}

// endregion

// region uhp::Processor::process
func (proc *uhpProcessor) process(received *uhpMessage) (optSend *uhpMessage) {
	switch {
	case received == nil:
		return nil
	case received.ENVELOPE != nil:
		return proc.processEnvelope(*received.ENVELOPE)
	case received.BROADCAST != nil:
		// log.Printf(LOG_PREFIX+"[BROADCAST:%d] received. processor: %#v", &proc.Id, &proc.Dir, *received.BROADCAST, proc)
		return proc.processBroadcast(*received.BROADCAST)
	case proc.Status == Active:
		return &uhpMessage{}
	default:
		return nil
	}
}

func (proc *uhpProcessor) processEnvelope(curr uhpEnvelope) *uhpMessage {
	if curr.Max < proc.Id {
		curr.Max = proc.Id
	}
	// log.Printf(LOG_PREFIX+"Received envelope %#v", proc.Id, proc.Dir, curr)
	if proc.testLeader(curr) {
		proc.Status = TempLeader
		log.Printf(LOG_PREFIX+"UHP found one of two leaders: %#v", proc.Id, proc.Dir, proc)
		return &uhpMessage{BROADCAST: &curr.Max}
	}

	// log.Printf(LOG_PREFIX+"comparing... [%#v] vs [%#v]", proc.Id, proc.Dir, proc.Last, curr)
	if drop, promote := proc.Promoter.Test(proc.Last, curr); !drop {
		if promote {
			log.Printf(LOG_PREFIX+"envelope promoted!", proc.Id, proc.Dir)
			proc.Promoter.Promote(&curr)
		}
		proc.Last = curr
		return &uhpMessage{ENVELOPE: &curr}
	} else {
		log.Printf(LOG_PREFIX+"dropping envelope!", proc.Id, proc.Dir)
		return &uhpMessage{}
	}
}

func (proc *uhpProcessor) processBroadcast(curr uhpBroadcast) (toSend *uhpMessage) {
	switch proc.Status {
	case Active:
		toSend = &uhpMessage{BROADCAST: &curr}
	case TempLeader:
		log.Printf(LOG_PREFIX+"[BROADCAST:%d] performed round trip. demoting UHP leader...", proc.Id, proc.Dir, curr)
		toSend = nil
	default:
		log.Panicf(LOG_PREFIX+"[BROADCAST:%d] invalid state received broadcast", proc.Id, proc.Dir, curr)
	}

	if proc.Id == curr {
		proc.Status = TrueLeader
		log.Printf(LOG_PREFIX+"[BROADCAST:%d] processor is true leader", proc.Id, proc.Dir, curr)
	} else {
		proc.Status = Dead
	}
	return toSend
}

// endregion

// endregion

// region driver
func run(v lib.Node) {
	v.StartProcessing()
	uhp{}.initiateElection(v)
	v.FinishProcessing(false)

	for _round, finished := 1, false; !finished; _round++ {
		v.StartProcessing()
		finished = uhp{}.round(v)
		v.FinishProcessing(finished)
	}
}

func verifyLeaderElected(vertices []lib.Node) {
	var leaderNode lib.Node
	var s uhpState
	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if s.IsLeader {
			leaderNode = v
			break
		}
	}
	if leaderNode == nil {
		panic("There is no leader on the undirected undirected_ring")
	}
	max := 0
	for _, v := range vertices {
		json.Unmarshal(v.GetState(), &s)
		if v.GetIndex() > max {
			max = v.GetIndex()
		}
		if v != leaderNode {
			if s.IsLeader {
				panic(fmt.Sprint(
					"Multiple leaders on the undirected undirected_ring: ", v.GetIndex(), leaderNode.GetIndex()))
			}
			// if s.Status != Processor {
			// 	panic(fmt.Sprint("Node ", v.GetIndex(), " has state ", s.Status))
			// }
		}
	}
	if max != leaderNode.GetIndex() {
		panic(fmt.Sprint("Leader has value ", leaderNode.GetIndex(), " but max is ", max))
	}
	log.Println("LEADER ", leaderNode.GetIndex())
}

func Run(n int) (int, int) {
	vertices, synchronizer := lib.BuildSynchronizedRing(n)
	for _, v := range vertices {
		go run(v)
	}
	synchronizer.Synchronize(0)
	verifyLeaderElected(vertices)
	return synchronizer.GetStats()
}

// endregion
