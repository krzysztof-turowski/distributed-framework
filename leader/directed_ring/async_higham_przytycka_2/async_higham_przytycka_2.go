// Algorithm ELECT from: L. Higham, T. Przytycka,
// "A simple, efficient algorithm for maximum finding on rings"
package async_higham_przytycka_2

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/krzysztof-turowski/distributed-framework/leader/directed_ring/common"
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

// Delay is the maximum random pause before a node handles a message.
var Delay time.Duration

type Algorithm int

const (
	Basic Algorithm = iota
	Elect
	FullElect // Elect with improvements in every round (not depending on the parity of the round)
)

type envelope struct {
	Id      int
	Round   int
	Counter int
	Elected bool
}

type state struct {
	Id       int
	FwdId    int
	FwdRound int
	MaxRound int
	Status   common.ModeType
	LeaderId int
}

func (s *state) isLeader(e envelope) bool {
	return e.Id == s.FwdId && e.Round == s.FwdRound
}

func (s *state) isCasualty(e envelope) bool {
	if e.Round != s.FwdRound {
		return false
	}
	if isOdd(e.Round) {
		return e.Id > s.FwdId
	}
	return e.Id < s.FwdId
}

func (s *state) isPromoted(e envelope, alg Algorithm) bool {
	if e.Round == 0 {
		return true
	}

	sameRound := e.Round == s.FwdRound
	witnessOdd := isOdd(e.Round) && e.Round == s.FwdRound+1 && e.Id < s.FwdId
	witnessEven := !isOdd(e.Round) && e.Round == s.FwdRound+1 && e.Id > s.FwdId

	switch alg {
	case Basic:
		return sameRound
	case Elect:
		if isOdd(e.Round) {
			return e.Counter == 0 || (sameRound && e.Id < s.FwdId)
		}
		return witnessEven
	case FullElect:
		return e.Counter == 0 || sameRound || witnessOdd || witnessEven
	default:
		return false
	}
}

// handle applies one step of the chosen algorithm to a received envelope.
// It reports whether the node still expects more messages.
func (s *state) handle(v lib.Node, e envelope, alg Algorithm) bool {
	switch {
	case e.Elected:
		s.LeaderId = e.Id
		if s.Status == common.Leader {
			return false
		}
		s.Status = common.NonLeader
		send(v, e)
		return false

	case s.isLeader(e):
		log.Printf("Node %d is the leader (id %d)", v.GetIndex(), s.Id)
		s.Status = common.Leader
		send(v, envelope{Id: s.Id, Elected: true})
		return true

	case s.isCasualty(e):
		log.Printf("Node %d destroys %+v", v.GetIndex(), e)
		return true

	default:
		if s.isPromoted(e, alg) {
			e.Round++
			if alg != Basic {
				e.Counter = fibonacci(e.Round + 2)
			}
		}
		s.MaxRound = max(s.MaxRound, e.Round) // for statistics only
		s.FwdId, s.FwdRound = e.Id, e.Round
		e.Counter--
		send(v, e)
		return true
	}
}

func initialize(v lib.Node, id int, alg Algorithm) {
	s := state{Id: id, FwdId: math.MinInt, FwdRound: -1, Status: common.Unknown}
	s.handle(v, envelope{Id: id, Round: 0}, alg)
	setState(v, s)
}

func process(v lib.Node, alg Algorithm) bool {
	randomPause()
	s := getState(v)
	e := receive(v)
	running := s.handle(v, e, alg)
	setState(v, s)
	return running
}

func run(v lib.Node, id int, alg Algorithm) {
	v.StartProcessing()
	initialize(v, id, alg)
	for process(v, alg) {
	}
	v.FinishProcessing(true)
}

func send(v lib.Node, e envelope) {
	log.Printf("Node %d sends %+v to node %d", v.GetIndex(), e, v.GetOutNeighbors()[0].GetIndex())
	data, _ := json.Marshal(e)
	v.SendMessage(0, data)
}

func receive(v lib.Node) envelope {
	var e envelope
	json.Unmarshal(v.ReceiveMessage(0), &e)
	return e
}

func getState(v lib.Node) state {
	var s state
	json.Unmarshal(v.GetState(), &s)
	return s
}

func setState(v lib.Node, s state) {
	data, _ := json.Marshal(s)
	v.SetState(data)
}

func randomPause() {
	if Delay > 0 {
		time.Sleep(time.Duration(rand.Int63n(int64(Delay))))
	}
}

func isOdd(n int) bool {
	return n%2 == 1
}

func fibonacci(n int) int {
	a, b := 0, 1
	for i := 0; i < n; i++ {
		a, b = b, a+b
	}
	return a
}

// check verifies that exactly one node is the leader, all nodes know its id,
// and returns the maximum algorithm round reached.
func check(vertices []lib.Node) int {
	var leader *state
	maxRound := 0
	for _, v := range vertices {
		s := getState(v)
		if s.MaxRound > maxRound {
			maxRound = s.MaxRound
		}
		if s.Status == common.Leader {
			if leader != nil {
				panic("more than one leader")
			}
			leader = &s
		} else if s.Status != common.NonLeader {
			panic(fmt.Sprintf("node %d finished with status %q", v.GetIndex(), s.Status))
		}
	}
	if leader == nil {
		panic("no leader")
	}
	for _, v := range vertices {
		if getState(v).LeaderId != leader.Id {
			panic(fmt.Sprintf("node %d does not know the leader", v.GetIndex()))
		}
	}
	log.Printf("Leader elected: id %d (maxRound: %d)", leader.Id, maxRound)
	return maxRound
}

// RunAlgorithm runs the chosen algorithm with the given ids
func RunAlgorithm(alg Algorithm, ids []int) (messages int, maxRound int, timeMs int) {
	vertices, runner := lib.BuildDirectedRing(len(ids))

	for i, v := range vertices {
		go run(v, ids[i], alg)
	}

	runner.Run(false)
	maxRound = check(vertices)
	messages, timeMs = runner.GetStats()
	return messages, maxRound, timeMs
}

// Run runs the Elect algorithm with random ids
func Run(n int) (int, int) {
	messages, _, timeMs := RunAlgorithm(Elect, rand.Perm(n))
	return messages, timeMs
}

// RunBasic runs the Basic algorithm with random ids
func RunBasic(n int) (int, int) {
	messages, _, timeMs := RunAlgorithm(Basic, rand.Perm(n))
	return messages, timeMs
}

// RunFullElect runs the FullElect algorithm with random ids
func RunFullElect(n int) (int, int) {
	messages, _, timeMs := RunAlgorithm(FullElect, rand.Perm(n))
	return messages, timeMs
}
