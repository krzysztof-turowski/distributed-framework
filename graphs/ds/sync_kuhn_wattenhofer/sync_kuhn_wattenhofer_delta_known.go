package sync_kuhn_wattenhofer

import (
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"math"
)

func RunWithMaxDegree(vertices []lib.Node, synchronizer lib.Synchronizer, roundsParam int) {
	if roundsParam <= 0 { panic("roundsParam has to be a positive integer") }
	
	delta := -1
	for _, v := range vertices {
		if v.GetOutChannelsCount() >= delta { delta = v.GetOutChannelsCount() }
	}

	init := func(v lib.Node) bool { return initializeWithDelta(v, roundsParam, delta) }
	runNode := func(v lib.Node) { run(v, init, processWhenDeltaIsKnown) }
	check := func(vertices []lib.Node) { checkDS(vertices, roundsParam) }
	runSynchronized(vertices, synchronizer, runNode, check)
}

type lpStateDelta struct {
	lpState
	Delta	int
}

func initializeWithDelta(v lib.Node, roundsParam, delta int) (finished bool) {
	setState(v, lpStateDelta{ createInitialLpState(roundsParam), delta })
	return false
}

func processWhenDeltaIsKnown(v lib.Node) (finished bool) {
	state := getState[lpStateDelta](v)
	if state.OuterIndex >= 0 {
		dynamicDegree := countTrueValuesInNeighborhood(v, state.Status == uncovered)
		degreeLowBound := math.Pow(float64(state.Delta + 1), float64(state.OuterIndex) / float64(state.RoundsParam))
		if float64(dynamicDegree) >= degreeLowBound {
			considered := math.Pow(float64(state.Delta + 1), -float64(state.InnerIndex) / float64(state.RoundsParam))
			if state.DominationValue < considered { state.DominationValue = considered }
		}
		
		if isCovered(v, state.DominationValue) { state.Status = covered }

		decreaseIndices(&state.lpState)
		setState(v, state)
		return false
	} else {
		state.Status = lpToIpMDSTransformation(v, state.DominationValue)
		setState(v, state.lpState)
		return true
	}
}
