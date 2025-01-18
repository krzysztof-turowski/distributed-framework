package sync_karp_widgerson

import (
	"encoding/json"
	"fmt"
	"math"
	"math/bits"
	"math/rand"

	"github.com/krzysztof-turowski/distributed-framework/lib"
)

type phaseType int

const (
	initPhase phaseType = iota
	heavyFindPhase
	scoreFindPhase
	indFindPhase
	finalPhase
)

type phaseStage int

const (
	indFindSendInPhaseSet phaseStage = iota
	indFindReceiveInPhaseSet
	indFindSendIsRemoved
	indFindReceiveIsRemoved
	indFindSendInMIS
	indFindReceiveInMIS
	indFindSendIsActive
	indFindReceiveIsActive

	heavyFindSendStatus
	heavyFindReceiveStatus
	heavyFindNotifyLeader
	heavyFindHandleDecision

	findLeaderSend
	findLeaderReceive
	findLeaderNotify
	findLeaderHandleNotifies

	scoreFindStart
	scoreFindPushResults
)

// Extend state with fields necessary for algorithm logic.
type state struct {
	Phase      phaseType
	PhaseStage phaseStage
	Active     bool
	InPhaseSet bool
	InMIS      bool

	OutIdByNeighbor map[int]int
	OutNeghbourId   []int
	IsNeigbourAlive []bool
	NeighbourInMIS  bool

	MaxId              int // For FINDLEADER: maximum ID in the graph
	RoundsLeftLeader   int // For FINDLEADER: rounds left to be sure MaxId is leader
	SenderOfMaxId      int // Id, to towards which there is sender
	PassMsgFromLeader  []bool
	DistanceFromLeader int

	Degree              int // For HEAVYFIND: current value of i
	LeaderIterations    int
	HSize               int
	CurrentI            int
	ShouldTerminate     bool
	CntOfAboveThreshold int
	LeaderThreshold     int

	// Leader communications helpers
	PushUpLeaderRound   int
	PushDownLeaderRound int

	VertexIsInT map[int]struct{} // For SCOREFIND: map of vertices in set T

	IsRemoved             bool   // For INDFIND: flag for removal from set T
	ShouldRemoveNeighbour []bool // FOR INDFIND: decisions if if we should remove neighbour

}

type message struct {
	Value    []byte
	SenderID int
}

func setState(v lib.Node, s state) {
	data, _ := json.Marshal(s)
	v.SetState(data)
}

func getState(v lib.Node) state {
	data := v.GetState()
	var s state
	json.Unmarshal(data, &s)
	return s
}

func sendMessage(v lib.Node, target int, msg message) {
	data, _ := json.Marshal(msg)
	v.SendMessage(target, data)
}

func receiveMessage(v lib.Node, source int) (message, bool) {
	data := v.ReceiveMessage(source)
	if data == nil {
		return message{}, false
	}
	var msg message
	json.Unmarshal(data, &msg)
	return msg, true
}

func bytesToBool(bytes []byte) bool {
	if (len(bytes) != 1) || (bytes[0] != 0 && bytes[0] != 1) {
		panic("Invalid byte value")
	}
	return bytes[0] != 0
}

func boolToBytes(b bool) []byte {
	result := make([]byte, 1)
	if b {
		result[0] = 1
	} else {
		result[0] = 0
	}
	return result
}

func initialize(v lib.Node) {
	outIdByNeighbor := make(map[int]int)
	outNeghbourId := make([]int, 0, v.GetOutChannelsCount())
	isNeigbourAlive := make([]bool, v.GetOutChannelsCount())
	passMsgFromLeader := make([]bool, v.GetOutChannelsCount())

	for i := 0; i < v.GetOutChannelsCount(); i++ {
		neighbour := v.GetOutNeighbors()[i]
		outIdByNeighbor[neighbour.GetIndex()] = i
		outNeghbourId = append(outNeghbourId, neighbour.GetIndex())
		isNeigbourAlive[i] = true
		passMsgFromLeader[i] = false
	}

	s := state{
		Phase:              initPhase,
		PhaseStage:         findLeaderSend,
		Active:             true,
		InPhaseSet:         true,
		InMIS:              false,
		NeighbourInMIS:     false,
		OutIdByNeighbor:    outIdByNeighbor,
		OutNeghbourId:      outNeghbourId,
		RoundsLeftLeader:   v.GetSize(),
		MaxId:              v.GetIndex(),
		PassMsgFromLeader:  passMsgFromLeader,
		LeaderIterations:   0,
		DistanceFromLeader: 0,

		PushDownLeaderRound: 0,
		PushUpLeaderRound:   0,
		HSize:               -1,
		IsNeigbourAlive:     isNeigbourAlive,
		SenderOfMaxId:       -1,
	}
	setState(v, s)
}

func sendToAliveNeighbours(v lib.Node, s *state, createMessage func(v lib.Node, s *state, neighbourId int) message) {
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		if s.IsNeigbourAlive[i] {
			sendMessage(v, i, createMessage(v, s, i))
		}
	}
}

func receiveFromAliveNeighbours(v lib.Node, s *state, handleMessage func(msg message, neighbourId int)) {
	for i := 0; i < v.GetOutChannelsCount(); i++ {
		if s.IsNeigbourAlive[i] {
			if msg, ok := receiveMessage(v, i); ok {
				handleMessage(msg, i)
			}
		}
	}
}

func pushUpLeader(v lib.Node, s *state, initInfo func(v lib.Node, s *state) []byte, concatChildInfo func([]byte, message) []byte, onLeader func(lib.Node, *state, []byte)) bool {
	currentHeight := v.GetSize() - s.PushUpLeaderRound - 1

	if currentHeight == s.DistanceFromLeader {

		messageValue := initInfo(v, s)

		for i := 0; i < v.GetInChannelsCount(); i++ {
			if s.PassMsgFromLeader[i] {
				if msg, ok := receiveMessage(v, i); ok {
					messageValue = concatChildInfo(messageValue, msg)
				} else if !ok {
					panic("Couldn't receive message")
				}
			}
		}

		if s.MaxId == v.GetIndex() {
			if currentHeight != 0 {
				panic("Leader should be at height 0")
			}
			onLeader(v, s, messageValue)
		} else {
			if s.SenderOfMaxId == -1 {
				panic("SenderOfMaxId should be set")
			}
			if _, exists := s.OutIdByNeighbor[s.SenderOfMaxId]; !exists {
				panic("SenderOfMaxId should be in OutIdByNeighbor")
			}
			sendMessage(v, s.OutIdByNeighbor[s.SenderOfMaxId], message{Value: messageValue, SenderID: v.GetIndex()})
		}
	}

	s.PushUpLeaderRound += 1
	if s.PushUpLeaderRound == v.GetSize() {
		s.PushUpLeaderRound = 0
		return true
	}

	return false
}

func pushDownLeader(v lib.Node, s *state, onParentInfo func(lib.Node, *state, message) []byte, onLeader func(lib.Node, *state) []byte) bool {
	currentHeight := s.PushDownLeaderRound
	if currentHeight == s.DistanceFromLeader {
		var messageValue []byte
		if s.MaxId == v.GetIndex() { // Leader
			if currentHeight != 0 {
				panic("Leader should be at height 0")
			}
			messageValue = onLeader(v, s)
		} else { // Not leader
			if msg, ok := receiveMessage(v, s.OutIdByNeighbor[s.SenderOfMaxId]); ok {
				messageValue = onParentInfo(v, s, msg)
			}
		}

		for i := 0; i < v.GetOutChannelsCount(); i++ {
			if s.PassMsgFromLeader[i] {
				sendMessage(v, i, message{Value: messageValue, SenderID: v.GetIndex()})
			}
		}
	}

	s.PushDownLeaderRound += 1

	if s.PushDownLeaderRound == v.GetSize() {
		s.PushDownLeaderRound = 0
		return true
	}

	return false
}

func processFindLeader(v lib.Node, s *state) {
	type LeaderMessage struct {
		MaxId              int
		DistanceFromLeader int
	}

	if s.PhaseStage == findLeaderSend {
		sendToAliveNeighbours(v, s, func(v lib.Node, s *state, neighbourId int) message {
			var leaderMessage LeaderMessage
			leaderMessage.MaxId = s.MaxId
			leaderMessage.DistanceFromLeader = s.DistanceFromLeader
			bytesToSend, _ := json.Marshal(leaderMessage)
			return message{Value: bytesToSend, SenderID: v.GetIndex()}
		})
		s.PhaseStage = findLeaderReceive
	} else if s.PhaseStage == findLeaderReceive {
		receiveFromAliveNeighbours(v, s,
			func(msg message, neighbourId int) {
				var leaderMessage LeaderMessage
				json.Unmarshal(msg.Value, &leaderMessage)
				if leaderMessage.MaxId > s.MaxId {
					s.MaxId = leaderMessage.MaxId
					s.DistanceFromLeader = leaderMessage.DistanceFromLeader + 1
					s.SenderOfMaxId = msg.SenderID
				}
			})
		s.RoundsLeftLeader -= 1
		if s.RoundsLeftLeader == 0 {
			s.PhaseStage = findLeaderNotify
		} else {
			s.PhaseStage = findLeaderSend
		}
	} else if s.PhaseStage == findLeaderNotify {
		sendToAliveNeighbours(v, s, func(v lib.Node, s *state, neighbourId int) message {
			return message{Value: boolToBytes(s.OutNeghbourId[neighbourId] == s.SenderOfMaxId), SenderID: v.GetIndex()}
		})

		foundSomebody := false
		for i := 0; i < v.GetOutChannelsCount(); i++ {
			if s.OutNeghbourId[i] == s.SenderOfMaxId {
				foundSomebody = true
			}
		}

		if !foundSomebody && !(v.GetIndex() == s.MaxId) {
			panic(fmt.Sprintf("Couldn't find from leader nodeId:%d, maxId:%d", v.GetIndex(), s.MaxId))
		}
		s.PhaseStage = findLeaderHandleNotifies
	} else if s.PhaseStage == findLeaderHandleNotifies {
		receiveFromAliveNeighbours(v, s,
			func(msg message, neighbourId int) {
				if bytesToBool(msg.Value) {
					s.PassMsgFromLeader[neighbourId] = true
				}
			})

		s.ShouldTerminate = false
		s.PhaseStage = heavyFindSendStatus
		s.Phase = heavyFindPhase
	}
}

func processHeavyFind(v lib.Node, s *state) {
	type HeavyLeaderNotifyMessage struct {
		VertexDegrees []int
	}

	type HeavyLeaderDecision struct {
		Threshold int
		Terminate bool
	}

	if s.PhaseStage == heavyFindSendStatus {
		sendToAliveNeighbours(v, s, func(v lib.Node, s *state, neighbourId int) message {
			return message{Value: boolToBytes(s.InPhaseSet), SenderID: v.GetIndex()}
		})

		s.PhaseStage = heavyFindReceiveStatus
	} else if s.PhaseStage == heavyFindReceiveStatus {
		s.Degree = 0

		receiveFromAliveNeighbours(v, s, func(msg message, neighbourId int) {
			if bytesToBool(msg.Value) {
				s.Degree += 1
			}
		})

		s.PhaseStage = heavyFindNotifyLeader
	} else if s.PhaseStage == heavyFindNotifyLeader {
		initInfo := func(v lib.Node, s *state) []byte {
			var heavyLeaderNotifyMessage HeavyLeaderNotifyMessage
			if s.InPhaseSet {
				heavyLeaderNotifyMessage.VertexDegrees = append(heavyLeaderNotifyMessage.VertexDegrees, s.Degree)
			}
			bytes, _ := json.Marshal(heavyLeaderNotifyMessage)
			return bytes
		}

		concatChildInfo := func(myInfo []byte, childMessage message) []byte {
			var heavyLeaderNotifyMessage HeavyLeaderNotifyMessage
			json.Unmarshal(childMessage.Value, &heavyLeaderNotifyMessage)

			var myHeavyLeaderNotifyMessage HeavyLeaderNotifyMessage
			json.Unmarshal(myInfo, &myHeavyLeaderNotifyMessage)

			myHeavyLeaderNotifyMessage.VertexDegrees = append(myHeavyLeaderNotifyMessage.VertexDegrees, heavyLeaderNotifyMessage.VertexDegrees...)
			bytes, _ := json.Marshal(myHeavyLeaderNotifyMessage)
			return bytes
		}

		onLeader := func(v lib.Node, s *state, messageValue []byte) {
			var heavyLeaderNotifyMessage HeavyLeaderNotifyMessage
			json.Unmarshal(messageValue, &heavyLeaderNotifyMessage)

			sizeOfK := len(heavyLeaderNotifyMessage.VertexDegrees)

			if s.LeaderIterations == 0 {
				s.HSize = sizeOfK
				s.CurrentI = int(math.Ceil(math.Log2(float64(s.HSize))))
				if s.HSize == 1 {
					s.CurrentI = 1
				}
			}
			s.CurrentI -= 1

			threshold := int(math.Pow(2, float64(s.CurrentI))) - 1
			cntOfAboveThreshold := 0
			for _, degree := range heavyLeaderNotifyMessage.VertexDegrees {
				if degree >= threshold {
					cntOfAboveThreshold += 1
				}
			}

			s.LeaderThreshold = threshold
			s.CntOfAboveThreshold = cntOfAboveThreshold

			s.LeaderIterations += 1
		}

		finishPush := pushUpLeader(v, s, initInfo, concatChildInfo, onLeader)

		if finishPush {
			s.PhaseStage = heavyFindHandleDecision
		}

	} else if s.PhaseStage == heavyFindHandleDecision {
		onParentInfo := func(v lib.Node, s *state, parentMessage message) []byte {
			var heavyLeaderDecision HeavyLeaderDecision
			json.Unmarshal(parentMessage.Value, &heavyLeaderDecision)

			if s.InPhaseSet && !heavyLeaderDecision.Terminate && s.Degree >= heavyLeaderDecision.Threshold {
				s.InPhaseSet = false
			}

			if heavyLeaderDecision.Terminate {
				s.ShouldTerminate = true
			}

			return parentMessage.Value
		}

		onLeader := func(v lib.Node, s *state) []byte {
			var heavyLeaderDecision HeavyLeaderDecision
			s.ShouldTerminate = s.CntOfAboveThreshold >= s.LeaderThreshold
			heavyLeaderDecision.Threshold = s.LeaderThreshold
			heavyLeaderDecision.Terminate = s.ShouldTerminate

			bytes, _ := json.Marshal(heavyLeaderDecision)
			return bytes
		}

		finishPush := pushDownLeader(v, s, onParentInfo, onLeader)

		if finishPush {
			if s.ShouldTerminate {
				s.LeaderIterations = 0
				s.Phase = scoreFindPhase
				s.PhaseStage = scoreFindStart
			} else { // Continue while
				s.PhaseStage = heavyFindSendStatus
			}
		}
	}
}

func processScoreFind(v lib.Node, s *state) {
	type EdgeInfo struct {
		FromId    int
		ToId      int
		FromIsInK bool
	}

	type ScoreFindMessage struct {
		EdgeInfos []EdgeInfo
	}

	type ScoreFindLeaderDecision struct {
		VertexIsInT map[int]struct{}
	}

	if s.PhaseStage == scoreFindStart { // Leader has to know about each edge
		initInfo := func(v lib.Node, s *state) []byte {
			var scoreFindMessage ScoreFindMessage
			for i := 0; i < v.GetOutChannelsCount(); i++ {
				edgeInfo := EdgeInfo{
					FromId:    v.GetIndex(),
					ToId:      v.GetOutNeighbors()[i].GetIndex(),
					FromIsInK: s.InPhaseSet,
				}
				scoreFindMessage.EdgeInfos = append(scoreFindMessage.EdgeInfos, edgeInfo)
			}
			bytes, _ := json.Marshal(scoreFindMessage)
			return bytes
		}

		concatChildInfo := func(myInfo []byte, childMessage message) []byte {
			var scoreFindMessage ScoreFindMessage
			json.Unmarshal(childMessage.Value, &scoreFindMessage)

			var myscoreFindMessage ScoreFindMessage
			json.Unmarshal(myInfo, &myscoreFindMessage)

			myscoreFindMessage.EdgeInfos = append(myscoreFindMessage.EdgeInfos, scoreFindMessage.EdgeInfos...)

			bytes, _ := json.Marshal(myscoreFindMessage)
			return bytes
		}

		onLeader := func(v lib.Node, s *state, messageValue []byte) {
			var scoreFindMessage ScoreFindMessage
			json.Unmarshal(messageValue, &scoreFindMessage)

			isNodeInK := make(map[int]bool)
			adjacentNodes := make(map[int][]int)

			for _, edgeInfo := range scoreFindMessage.EdgeInfos {
				if edgeInfo.FromIsInK {
					isNodeInK[edgeInfo.FromId] = true
				}
			}

			for _, edgeInfo := range scoreFindMessage.EdgeInfos {
				if isNodeInK[edgeInfo.FromId] && isNodeInK[edgeInfo.ToId] {
					adjacentNodes[edgeInfo.FromId] = append(adjacentNodes[edgeInfo.FromId], edgeInfo.ToId)
				}
			}

			if s.HSize == 1 {
				s.VertexIsInT = make(map[int]struct{})
				for nodeId, status := range isNodeInK {
					if status {
						s.VertexIsInT[nodeId] = struct{}{}
					}
				}
				return
			}

			// l ← max{l′ | 2^l′ − 1 ∈ [1, (|H|/⌊log |H|⌋)]}
			l := 0
			for math.Pow(2, float64(l+1))-1 <= float64(s.HSize)/math.Floor(math.Log2(float64(s.HSize))) {
				l++

			}

			m := int(math.Pow(2, float64(l)) - 1)

			// Select M
			nodesInM := make([]int, 0)
			for nodeId, status := range isNodeInK {
				if status && len(nodesInM) != m {
					nodesInM = append(nodesInM, nodeId)
				}
			}

			if len(nodesInM) != m {
				panic("Couldn't find enough heavy verts")
			}

			delta := 0

			for _, adjNodes := range adjacentNodes {
				delta = int(math.Max(float64(delta), float64(len(adjNodes))))
			}

			// sc ← max{s′ | 2s′ − 1 ∈ [1, ⌈m/(16∆)⌉]}
			upperBound := float64(max(1, int(math.Ceil(float64(m)/(16*float64(delta))))))

			sc := 1
			for math.Pow(2, float64(sc+1))-1 <= upperBound {
				sc++
			}

			t := int(math.Max(2, float64(sc))) - 1

			type pair struct {
				from int
				to   int
			}

			profByNode := make(map[int]int)
			costByNode := make(map[pair]int)
			for nodeId, adjNodes := range adjacentNodes {
				profByNode[nodeId] = len(adjNodes) + 1

				adjSet := make(map[int]struct{})
				for _, node := range adjNodes {
					adjSet[node] = struct{}{}
				}

				for otherNodeId, otherAdjNodes := range adjacentNodes {
					if nodeId == otherNodeId {
						continue
					}
					if _, exists := adjSet[otherNodeId]; !exists {
						intersection := []int{}
						for _, intersectionId := range otherAdjNodes {
							if _, exists := adjSet[intersectionId]; exists {
								intersection = append(intersection, intersectionId)
							}
						}

						costByNode[pair{nodeId, otherNodeId}] = len(intersection)
					} else {
						costByNode[pair{nodeId, otherNodeId}] = max(profByNode[nodeId], profByNode[otherNodeId])
					}
				}
			}

			for j := l; j >= sc+1; j-- {
				//Construct a block design with set of elements U and parameters: v = 2j − 1, r = k = 2j−1 − 1, λ = 2j−2 − 1
				maxRating := float64(math.MinInt64)
				maxNodes := make([]int, 0)
				for mask := 1; mask < 1<<uint(j); mask++ {
					blockNodes := make([]int, 0, int(math.Pow(2, float64(j-1))-1))
					for otherMask := 1; otherMask < 1<<uint(j); otherMask++ {
						intersectionMask := mask & otherMask
						if bits.OnesCount(uint(intersectionMask))%2 == 0 {
							blockNodes = append(blockNodes, nodesInM[otherMask-1])
						}
					}

					// Compute rating
					rating := float64(0)
					for _, node := range blockNodes {
						rating += float64(t) / float64(len(nodesInM)) * float64(profByNode[node])
					}

					for i := 0; i < len(blockNodes); i++ {
						for j := i + 1; j < len(blockNodes); j++ {
							coeff := float64(t*(t+1)/2) / float64((len(blockNodes) * (len(blockNodes) - 1) / 2))
							rating -= coeff * float64(costByNode[pair{blockNodes[i], blockNodes[j]}])
						}
					}

					if rating > maxRating {
						maxRating = rating
						maxNodes = blockNodes
					}
				}

				if len(maxNodes) == 0 {
					panic("No nodes in maxNodes")
				}

				nodesInM = maxNodes
			}
			s.VertexIsInT = make(map[int]struct{})
			for _, node := range nodesInM {
				s.VertexIsInT[node] = struct{}{}
			}
		}

		pushFinish := pushUpLeader(v, s, initInfo, concatChildInfo, onLeader)

		if pushFinish {
			s.PhaseStage = scoreFindPushResults
		}
	} else {
		onParentInfo := func(v lib.Node, s *state, parentMessage message) []byte {
			var scoreFindLeaderDecision ScoreFindLeaderDecision
			json.Unmarshal(parentMessage.Value, &scoreFindLeaderDecision)

			if _, exists := scoreFindLeaderDecision.VertexIsInT[v.GetIndex()]; exists {
				s.InPhaseSet = true
			} else {
				s.InPhaseSet = false
			}

			return parentMessage.Value
		}

		onLeader := func(v lib.Node, s *state) []byte {
			var scoreFindLeaderDecision ScoreFindLeaderDecision
			scoreFindLeaderDecision.VertexIsInT = s.VertexIsInT

			bytes, _ := json.Marshal(scoreFindLeaderDecision)
			return bytes
		}

		finishPush := pushDownLeader(v, s, onParentInfo, onLeader)

		if finishPush {
			s.Phase = indFindPhase
			s.PhaseStage = indFindSendInPhaseSet
		}
	}
}

// Process INDFIND phase.
func processIndFind(v lib.Node, s *state) {
	type IndFindMessage struct {
		IsSenderInPhaseSet bool
		Flag               bool
	}

	if s.PhaseStage == indFindSendInPhaseSet {
		sendToAliveNeighbours(v, s, func(v lib.Node, s *state, neighbourId int) message {
			var indFindPhaseMessage IndFindMessage
			indFindPhaseMessage.IsSenderInPhaseSet = s.InPhaseSet
			indFindPhaseMessage.Flag = false
			bytesToSend, _ := json.Marshal(indFindPhaseMessage)

			return message{Value: bytesToSend, SenderID: v.GetIndex()}
		})
		s.PhaseStage = indFindReceiveInPhaseSet
	} else if s.PhaseStage == indFindReceiveInPhaseSet {
		s.ShouldRemoveNeighbour = make([]bool, v.GetInChannelsCount())
		for i := 0; i < v.GetInChannelsCount(); i++ {
			s.ShouldRemoveNeighbour[i] = false
		}
		receiveFromAliveNeighbours(v, s, func(msg message, neighbourId int) {
			shouldRemoveHim := false

			var indFindPhaseMessage IndFindMessage
			json.Unmarshal(msg.Value, &indFindPhaseMessage)

			if s.InPhaseSet && indFindPhaseMessage.IsSenderInPhaseSet {
				if msg.SenderID < v.GetIndex() {
					shouldRemoveHim = (rand.Intn(2) == 1)
					if !shouldRemoveHim {
						s.IsRemoved = true
					}
				}
			}

			s.ShouldRemoveNeighbour[neighbourId] = shouldRemoveHim
		})

		s.PhaseStage = indFindSendIsRemoved
	} else if s.PhaseStage == indFindSendIsRemoved {
		sendToAliveNeighbours(v, s, func(v lib.Node, s *state, neighbourId int) message {
			var indFindPhaseMessage IndFindMessage
			indFindPhaseMessage.Flag = s.ShouldRemoveNeighbour[neighbourId]
			indFindPhaseMessage.IsSenderInPhaseSet = s.InPhaseSet
			bytesToSend, _ := json.Marshal(indFindPhaseMessage)

			return message{Value: bytesToSend, SenderID: v.GetIndex()}
		})

		s.PhaseStage = indFindReceiveIsRemoved
	} else if s.PhaseStage == indFindReceiveIsRemoved {
		receiveFromAliveNeighbours(v, s, func(msg message, neighbourId int) {
			var indFindPhaseMessage IndFindMessage
			json.Unmarshal(msg.Value, &indFindPhaseMessage)
			if indFindPhaseMessage.Flag { // If removed
				s.IsRemoved = true
			}
		})

		s.PhaseStage = indFindSendInMIS
	} else if s.PhaseStage == indFindSendInMIS {

		if !s.IsRemoved && s.InPhaseSet {
			if !s.Active {
				panic("Not active")
			}
			s.InMIS = true
		}

		sendToAliveNeighbours(v, s, func(v lib.Node, s *state, neighbourId int) message {
			var indFindPhaseMessage IndFindMessage
			indFindPhaseMessage.IsSenderInPhaseSet = s.InPhaseSet
			indFindPhaseMessage.Flag = s.InMIS
			bytesToSend, _ := json.Marshal(indFindPhaseMessage)

			return message{Value: bytesToSend, SenderID: v.GetIndex()}
		})

		s.PhaseStage = indFindReceiveInMIS
	} else if s.PhaseStage == indFindReceiveInMIS {
		receiveFromAliveNeighbours(v, s, func(msg message, neighbourId int) {
			var indFindPhaseMessage IndFindMessage
			json.Unmarshal(msg.Value, &indFindPhaseMessage)
			if indFindPhaseMessage.Flag {
				s.NeighbourInMIS = true
			}
		})

		s.PhaseStage = indFindSendIsActive
	} else if s.PhaseStage == indFindSendIsActive {
		sendToAliveNeighbours(v, s, func(v lib.Node, s *state, neighbourId int) message {
			var indFindPhaseMessage IndFindMessage
			indFindPhaseMessage.IsSenderInPhaseSet = s.InPhaseSet
			indFindPhaseMessage.Flag = s.Active && (s.InMIS || s.NeighbourInMIS)
			bytesToSend, _ := json.Marshal(indFindPhaseMessage)
			return message{Value: bytesToSend, SenderID: v.GetIndex()}
		})

		s.PhaseStage = indFindReceiveIsActive
	} else if s.PhaseStage == indFindReceiveIsActive {
		receiveFromAliveNeighbours(v, s, func(msg message, neighbourId int) {
			var indFindPhaseMessage IndFindMessage
			json.Unmarshal(msg.Value, &indFindPhaseMessage)
			if indFindPhaseMessage.Flag { // It was active, but will be removed now
				s.IsNeigbourAlive[neighbourId] = false
			}
		})

		if s.Active && (s.InMIS || s.NeighbourInMIS) { // Finishing execution
			s.Active = false
			s.Phase = finalPhase
		} else {

			s.MaxId = v.GetIndex()
			s.IsRemoved = false
			s.InPhaseSet = true
			s.Phase = initPhase
			s.PhaseStage = findLeaderSend
			s.DistanceFromLeader = 0
			s.RoundsLeftLeader = v.GetSize()
			s.SenderOfMaxId = -1
			for i := 0; i < v.GetOutChannelsCount(); i++ {
				s.PassMsgFromLeader[i] = false
			}
		}
	}
}

func process(v lib.Node) bool {
	s := getState(v)
	switch s.Phase {
	case initPhase:
		processFindLeader(v, &s)
	case heavyFindPhase:
		processHeavyFind(v, &s)
	case scoreFindPhase:
		processScoreFind(v, &s)
	case indFindPhase:
		processIndFind(v, &s)
	case finalPhase:
		setState(v, s)
		return true
	}

	setState(v, s)
	return false
}

func run(v lib.Node) {
	initialize(v)
	finished := false
	for !finished {
		v.StartProcessing()
		finished = process(v)
		v.FinishProcessing(finished)
	}
}

func check(vertices []lib.Node) {
	for _, v := range vertices {
		s := getState(v)

		if s.InMIS && s.NeighbourInMIS {
			panic("Vertex has flag in MIS and NeighbourInMIS")
		}

		if s.InMIS {
			for i := 0; i < v.GetOutChannelsCount(); i++ {
				neighborState := getState(v.GetOutNeighbors()[i])
				if neighborState.InMIS {
					panic("Two neighbors are in MIS")
				}
			}
		} else {
			covered := false
			for i := 0; i < v.GetOutChannelsCount(); i++ {
				neighborState := getState(v.GetOutNeighbors()[i])
				covered = covered || neighborState.InMIS
			}
			if !covered {
				panic("Not Maximal Independent Set, vertex is not covered and is not in MIS")
			}
		}
	}
}

func Run(n int, p float64) (int, int) {
	vertices, synchronizer := lib.BuildSynchronizedRandomGraph(n, p)
	for _, v := range vertices {
		go run(v)
	}
	synchronizer.Synchronize(0)
	check(vertices)
	return synchronizer.GetStats()
}
