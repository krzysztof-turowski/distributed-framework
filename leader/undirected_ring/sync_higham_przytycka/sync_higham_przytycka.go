package sync_higham_przytycka

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

const DEBUG_ID_MOD = 1000

type ID = int
type direction byte

const (
	Left  direction = iota
	Right direction = iota
)

// region message types
type uhpEnvelope struct {
	Id    ID
	Dir   direction
	Round uint
	Cnt   int
	Max   ID
}

type uhpBroadcast = ID

type uhpMessageType byte

const (
	UHP_T_ENVELOPE  uhpMessageType = iota
	UHP_T_BROADCAST uhpMessageType = iota
)

type uhpMessage struct {
	ENVELOPE  *uhpEnvelope
	BROADCAST *uhpBroadcast
}
type uhpMailbox = uhpMessage

// endregion

// region state struct
type undirectedHighamPrzytycka struct {
	IsLeader           bool // dependent on both processors
	ReceivedBroadcasts [2]bool
	Processors         [2]uhpProcessor
}

// used as namespace
type uhp = undirectedHighamPrzytycka
type uhpState = uhp

// endregion

// region utility types
type uhpSender = func([]byte)
type uhpReceiver = func() []byte // receiving as a means of creating a new state is easier than generics
type uhpCallbacks struct {
	send    uhpSender
	receive uhpReceiver
}

type uhpProcessor struct {
	Mailbox   *uhpMailbox
	Callbacks uhpCallbacks

	Id       ID
	Dir      direction // for observability
	Last     uhpEnvelope
	IsLeader bool
}

// endregion

// region uhp:: (scope: bidirectional)

// needed this function because closures aren't serializable
// and won't be preserved when saving state to JSON
func (uhp) prepareCallbacks(state *uhpState, node lib.Node) {
	senderBuilder := func(d direction) uhpSender {
		return func(payload []byte) {
			node.SendMessage(ID(d), payload)
		}
	}

	receiverBuilder := func(d direction) uhpReceiver {
		return func() []byte {
			return node.ReceiveMessage(int(d))
		}
	}

	rtlCallbacks := uhpCallbacks{
		send:    senderBuilder(Left),
		receive: receiverBuilder(Right),
	}
	ltrCallbacks := uhpCallbacks{
		send:    senderBuilder(Right),
		receive: receiverBuilder(Left),
	}

	state.Processors[Left].Callbacks = ltrCallbacks
	state.Processors[Right].Callbacks = rtlCallbacks
}

func (uhp) getState(node lib.Node) (state uhpState) {
	json.Unmarshal(node.GetState(), &state)
	uhp{}.prepareCallbacks(&state, node)
	return state
}

func (uhp) updateState(node lib.Node, state uhpState) {
	if payload, err := json.Marshal(state); err == nil {
		node.SetState(payload)
	} else {
		panic(err)
	}
}

// region uhp::constructors
func (uhp) newInstance(node lib.Node) uhpState {
	// assuming ID type doesn't overflow so processors have different IDs,
	// but it is not required for correctness
	id := node.GetIndex()
	res := uhpState{
		IsLeader:           false,
		ReceivedBroadcasts: [2]bool{},
		Processors: [2]uhpProcessor{
			uhp{}.newProcessor(id, Left),
			uhp{}.newProcessor(id, Right),
		},
	}
	uhp{}.prepareCallbacks(&res, node)
	return res
}

func (uhp) newEnvelope(id int, d direction) *uhpMessage {
	// id = id % DEBUG_ID_MOD
	log.Printf("Created envelope with id: %+04d", id)
	return &uhpMessage{
		ENVELOPE: &uhpEnvelope{
			Id:    id,
			Dir:   d,
			Round: 0,
			Max:   id,
		},
	}
}

func (uhp) newProcessor(id int, d direction) uhpProcessor {
	return uhpProcessor{
		Id:       id,
		Dir:      d,
		IsLeader: false,
		Mailbox:  uhp{}.newEnvelope(id, d),
	}
}

// endregion

func (uhp) initiateElection(node lib.Node) {
	state := uhp{}.newInstance(node)

	var wg sync.WaitGroup
	singlePass := func(p *uhpProcessor) {
		defer wg.Done()
		go p.send(p.Mailbox)
		p.Last = *p.Mailbox.ENVELOPE
		p.Mailbox = p.receive()
	}

	wg.Add(2)
	go singlePass(&state.Processors[Left])
	go singlePass(&state.Processors[Right])
	wg.Wait()

	uhp{}.updateState(node, state)
}

func (uhp) round(node lib.Node) (finished bool) {
	state := uhp{}.getState(node)

	results := make(chan bool)
	runRound := func(proc *uhpProcessor) {
		rc := &state.ReceivedBroadcasts[proc.Dir]
		if *rc {
			results <- true
		} else {
			*rc = proc.process()
			results <- *rc
		}
	}

	go runRound(&state.Processors[Left])
	go runRound(&state.Processors[Right])

	<-results
	<-results

	testTrueLeader := func(proc uhpProcessor) bool {
		return proc.IsLeader
	}
	if state.ReceivedBroadcasts[Left] && state.ReceivedBroadcasts[Right] {
		state.IsLeader = testTrueLeader(state.Processors[Left]) &&
			testTrueLeader(state.Processors[Right])
		finished = true
	}

	// silly defer evaluation made me lose hours on this line
	uhp{}.updateState(node, state)
	return finished
}

// endregion

// region uhp::Processor (scope: unidirectional)
// region uhp::Processor::test (TODO)
func (proc uhpProcessor) testEnvelope(curr uhpEnvelope) (drop bool, promote bool) {
	last := proc.Last
	drop = curr.Round == last.Round // equal rounds are a necessary condition to drop envelope
	switch curr.Round % 2 {
	case 0: // even
		drop = drop && curr.Id < last.Id
		promote = last.Round == curr.Round-1
	case 1: // odd
		drop = drop && curr.Id > last.Id
		promote = curr.Cnt == 0
	}
	promote = promote && !drop // promote implies !drop
	return drop, promote
	// return drop, promote
	// from the paper:
	// Casualty-test
	// return last.Round == curr.Round &&
	// 	((curr.Round%2 == 1 && curr.Id > last.Id) ||
	// 		(curr.Round%2 == 0 && curr.Id < last.Id))
	// Promotion-test:
	// (even(rnd) and fwd_round = rnd—1 and id > fwd_label) or
	// (odd(rnd) and ent= 0) or
	// (odd(rnd) and fwd_round = rnd and id < fwd_label).
}

func (proc uhpProcessor) testLeader(curr uhpEnvelope) bool {
	// envelope returned to owner
	return curr == proc.Last
}

// endregion

// region uhp::Processor::send
func (proc uhpProcessor) send(msg *uhpMessage) {
	if payload, err := json.Marshal(msg); err == nil {
		proc.Callbacks.send(payload)
	} else {
		panic("Failed to encode message")
	}
}

// endregion

// region uhp::Processor::receive
func (proc *uhpProcessor) receive() (result *uhpMessage) {
	payload := proc.Callbacks.receive()

	if err := json.Unmarshal(payload, &result); err == nil {
		return result
	} else {
		return &uhpMessage{}
	}
}

// endregion

// region uhp::Processor::process
func (proc *uhpProcessor) process() (receivedBroadcast bool) {
	switch {
	case proc.Mailbox == nil:
		log.Printf("Assuming empty mailbox...")
		go proc.send(nil)
	case proc.Mailbox.ENVELOPE != nil:
		proc.processEnvelope()
	case proc.Mailbox.BROADCAST != nil:
		proc.processBroadcast()
		log.Printf("RECEIVED BROADCAST: Processor %#v", proc)
		return true
	}

	proc.Mailbox = proc.receive()
	return false
}

func (proc *uhpProcessor) processEnvelope() {
	curr := *proc.Mailbox.ENVELOPE
	if curr.Max < proc.Id {
		curr.Max = proc.Id
	}

	// log.Printf("REceived envelope %#v", curr)
	if proc.testLeader(curr) {
		proc.IsLeader = true
		log.Printf("LEADER FOUND!!!!!!!!!!! %#v", proc)
		go proc.send(&uhpMessage{BROADCAST: &curr.Max})
	} else if drop, _ := proc.testEnvelope(curr); !drop {
		log.Printf("comparing... [%#v] vs [%#v] -> %#v", proc.Last, curr, drop)

		curr := curr
		if curr.Round == proc.Last.Round {
			curr.Round++
			// curr.cnt = F[rnd+2] __jm__ +2 to the new value?
		}

		proc.Last = curr
		go proc.send(&uhpMessage{ENVELOPE: &curr})
	} else {
		log.Printf("comparing... [%#v] vs [%#v] -> %#v", proc.Last, curr, drop)
		log.Printf("dropping envelope %#v!", curr)
		go proc.send(nil)
	}
}

func (proc *uhpProcessor) processBroadcast() {
	if proc.IsLeader {
		proc.IsLeader = false
	} else {
		proc.IsLeader = proc.Id == *proc.Mailbox.BROADCAST
		go proc.send(proc.Mailbox)
	}
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
		if _round == 50 {
			v.FinishProcessing(true)
			break
		}
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
