package sync_kuhn_wattenhofer

import (
	"github.com/krzysztof-turowski/distributed-framework/lib"
	"math"
)

func Run(vertices []lib.Node, synchronizer lib.Synchronizer, roundsParam int) {
	if roundsParam <= 0 { panic("roundsParam has to be a positive integer") }

	init := func(v lib.Node) bool { return initializeWithoutDelta(v, roundsParam) }
	runNode := func(v lib.Node) { run(v, init, processWhenDeltaIsUnknown) }
	check := func(vertices []lib.Node) { checkDS(vertices, roundsParam) }
	runSynchronized(vertices, synchronizer, runNode, check)
}

type lpStateDynamic struct {
	lpState
	DynamicDegree	int
	DynamicDegree2	int
}

func initializeWithoutDelta(v lib.Node, roundsParam int) bool {
	dynamicDegree := v.GetOutChannelsCount() + 1
	dynamicDegree2 := highestValueLocally(v, dynamicDegree, 2)
	setState(v, lpStateDynamic{ createInitialLpState(roundsParam), dynamicDegree, dynamicDegree2 })
	return false
}

func processWhenDeltaIsUnknown(v lib.Node) bool {
	state := getState[lpStateDynamic](v)

	if state.OuterIndex >= 0 {
		if state.DynamicDegree2 == undefined {
			state.DynamicDegree2 = highestValueLocally(v, state.DynamicDegree, 2)
		} else {
			lowBound := math.Pow(float64(state.DynamicDegree2), float64(state.OuterIndex) / float64(state.OuterIndex + 1))
			if lowBound == 0 { lowBound = 1 } //!!
			isActive := float64(state.DynamicDegree) >= lowBound
			
			activenessValue := countTrueValuesInNeighborhood(v, isActive)
			if state.Status == covered { activenessValue = 0 }
			activenessValue1 := highestValueLocally(v, activenessValue, 1)

			if isActive {
				considered := math.Pow(float64(activenessValue1), -float64(state.InnerIndex) / float64(state.InnerIndex + 1))
				if state.DominationValue < considered { state.DominationValue = considered }
			}

			if isCovered(v, state.DominationValue) { state.Status = covered }
			state.DynamicDegree = countTrueValuesInNeighborhood(v, state.Status == uncovered)

			if state.InnerIndex == 0 { state.DynamicDegree2 = undefined }
			decreaseIndices(&state.lpState)
		}
		setState(v, state)
		return false
	} else {
		state.Status = lpToIpMDSTransformation(v, state.DominationValue)
		setState(v, state.lpState)
		return true
	}
}
