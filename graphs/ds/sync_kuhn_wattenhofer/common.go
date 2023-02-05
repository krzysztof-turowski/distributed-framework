package sync_kuhn_wattenhofer

import (
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"encoding/json"
	"math"
	"log"
	"math/rand"
	"strconv"
)

func sendAll(v lib.Node, value interface{}) {
	msg, err := json.Marshal(value)
	if err != nil { panic(err) }
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		v.SendMessage(i, msg)
	}
}

func receiveAggregate[T interface{}](v lib.Node, aggregate func(nextValue T)) {
	var buffer T
	for i := 0; i < v.GetInChannelsCount(); i++ {
		msg := v.ReceiveMessage(i)
		err := json.Unmarshal(msg, &buffer)
		if err != nil { panic(err) }
		aggregate(buffer)
	}
}

const (
	undefined int = -1
)

type nodeStatus int

const (
    uncovered nodeStatus = iota
    covered

	outDS
	inDS
)

type lpState struct {
	RoundsParam		int
	OuterIndex		int
	InnerIndex		int

	DominationValue	float64

	Status			nodeStatus
}

func createInitialLpState(roundsParam int) lpState {
	return lpState{ roundsParam, roundsParam - 1, roundsParam - 1, 0.0, uncovered }
}

func getState[T interface{}](v lib.Node) T {
	var buffer T
	err := json.Unmarshal(v.GetState(), &buffer)
	if err != nil { panic(err) }
	return buffer
}

func setState[T interface{}](v lib.Node, state T) {
	serialized, err := json.Marshal(state)
	if err != nil { panic(err) }
	v.SetState(serialized)
}

func countTrueValuesInNeighborhood(v lib.Node, condition bool) int {
	sendAll(v, condition)
	truths := 0
	if condition { truths++ }
	receiveAggregate[bool](v, func(next bool) { if next { truths++ } })
	return truths
}

func highestValueLocally(v lib.Node, nodeValue int, maxDistance int) int {
	result := nodeValue
	for i := 0; i < maxDistance; i++ {
		sendAll(v, result)
		receiveAggregate[int](v, func(next int) { if result < next { result = next } })
	}
	return result
}

func isCovered(v lib.Node, dominationValue float64) bool {
	sendAll(v, dominationValue)
	coverage := dominationValue
	receiveAggregate[float64](v, func(next float64) { coverage += next })
	return coverage >= 1.0
}

func lpToIpMDSTransformation(v lib.Node, dominationValue float64) nodeStatus {
	delta2 := highestValueLocally(v, v.GetOutChannelsCount(), 2)
	probability := dominationValue * math.Log(float64(delta2 + 1))
	r := rand.Float64()
	log.Println("probability", probability, r)
	dominating := probability > r
	sendAll(v, dominating)
	dominated := dominating
	receiveAggregate[bool](v, func(next bool) { dominated = dominated || next })
	if !dominated { dominating = true }
	
	if dominating { return inDS } else { return outDS }
}

func checkDS(vertices []lib.Node, roundsParam int) {
	dsSize := 0
	for _, v := range vertices {
		state := getState[lpState](v)
		if state.Status == inDS {
			dsSize++	
		} else if state.Status == outDS {
			covered := false
			for _, neighbor := range v.GetOutNeighbors() {
				neighborState := getState[lpState](neighbor)
				covered = covered || (neighborState.Status == inDS)
			}
			if !covered { panic("no dominator in closed neighborhood") }
		} else {
			panic("invalid status " + strconv.Itoa(int(state.Status)))
		}
	}
	log.Println("size of dominating set =", dsSize)
}

func decreaseIndices(state *lpState) {
	state.InnerIndex--
	if state.InnerIndex < 0 {
		state.OuterIndex--
		state.InnerIndex += state.RoundsParam
	}	
}
