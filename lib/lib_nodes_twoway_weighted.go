package lib

import "math/rand"

type twoWayWeightedGraphNode struct {
	node    *twoWayNode
	weights []int
}

func (v *twoWayWeightedGraphNode) GetInWeights() []int {
	return v.weights
}

func (v *twoWayWeightedGraphNode) GetOutWeights() []int {
	return v.weights
}

func (v *twoWayWeightedGraphNode) ReceiveMessage(index int) []byte {
	return v.node.ReceiveMessage(index)
}

func (v *twoWayWeightedGraphNode) ReceiveAnyMessage() (int, []byte) {
	return v.node.ReceiveAnyMessage()
}

func (v *twoWayWeightedGraphNode) SendMessage(index int, message []byte) {
	v.node.SendMessage(index, message)
}

func (v *twoWayWeightedGraphNode) GetInChannelsCount() int {
	return v.node.GetInChannelsCount()
}

func (v *twoWayWeightedGraphNode) GetOutChannelsCount() int {
	return v.node.GetOutChannelsCount()
}

func (v *twoWayWeightedGraphNode) GetInNeighbors() []Node {
	return v.node.GetInNeighbors()
}

func (v *twoWayWeightedGraphNode) GetOutNeighbors() []Node {
	return v.node.GetOutNeighbors()
}

func (v *twoWayWeightedGraphNode) GetIndex() int {
	return v.node.GetIndex()
}

func (v *twoWayWeightedGraphNode) GetState() []byte {
	return v.node.GetState()
}

func (v *twoWayWeightedGraphNode) SetState(state []byte) {
	v.node.SetState(state)
}

func (v *twoWayWeightedGraphNode) GetSize() int {
	return v.node.GetSize()
}

func (v *twoWayWeightedGraphNode) StartProcessing() {
	v.node.StartProcessing()
}

func (v *twoWayWeightedGraphNode) FinishProcessing(finish bool) {
	v.node.FinishProcessing(finish)
}

func (v *twoWayWeightedGraphNode) shuffleTopology() {
	rand.Shuffle(len(v.node.neighborsChannels), func(i, j int) {
		v.node.neighborsChannels[i], v.node.neighborsChannels[j] = v.node.neighborsChannels[j], v.node.neighborsChannels[i]
		v.node.neighbors[i], v.node.neighbors[j] = v.node.neighbors[j], v.node.neighbors[i]
		v.node.neighborsCases[i], v.node.neighborsCases[j] = v.node.neighborsCases[j], v.node.neighborsCases[i]
		v.weights[i], v.weights[j] = v.weights[j], v.weights[i]
	})
}

func addTwoWayWeightedConnection(
	firstNode *twoWayWeightedGraphNode, secondNode *twoWayWeightedGraphNode,
	firstChan chan []byte, secondChan chan []byte, weight int) {
	addTwoWayConnection(firstNode.node, secondNode.node, firstChan, secondChan)
	firstNode.weights = append(firstNode.weights, weight)
	secondNode.weights = append(secondNode.weights, weight)
}
