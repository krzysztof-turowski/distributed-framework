package lib

import (
	"math/rand"
	"time"
)

type twoWayWeightedGraphNode struct {
	node    *oneWayNode
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

func (v *twoWayWeightedGraphNode) ReceiveMessageIfAvailable(index int) []byte {
	return v.node.ReceiveMessageIfAvailable(index)
}

func (v *twoWayWeightedGraphNode) ReceiveMessageWithTimeout(index int, timeout time.Duration) []byte {
	return v.node.ReceiveMessageWithTimeout(index, timeout)
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

func (v *twoWayWeightedGraphNode) IgnoreFutureMessages() {
	v.node.IgnoreFutureMessages()
}

func (v *twoWayWeightedGraphNode) Close() {
	v.node.Close()
}

func (v *twoWayWeightedGraphNode) shuffleTopology() {
	rand.Shuffle(len(v.node.inNeighborsChannels), func(i, j int) {
		v.node.inNeighborsChannels[i], v.node.inNeighborsChannels[j] = v.node.inNeighborsChannels[j], v.node.inNeighborsChannels[i]
		v.node.outNeighborsChannels[i], v.node.outNeighborsChannels[j] = v.node.outNeighborsChannels[j], v.node.outNeighborsChannels[i]
		v.node.inNeighbors[i], v.node.inNeighbors[j] = v.node.inNeighbors[j], v.node.inNeighbors[i]
		v.node.outNeighbors[i], v.node.outNeighbors[j] = v.node.outNeighbors[j], v.node.outNeighbors[i]
		v.node.inNeighborsCases[i], v.node.inNeighborsCases[j] = v.node.inNeighborsCases[j], v.node.inNeighborsCases[i]
		v.weights[i], v.weights[j] = v.weights[j], v.weights[i]
	})
}

func addTwoWayWeightedConnection(
	firstNode, secondNode *twoWayWeightedGraphNode, firstChan, secondChan chan []byte, weight int) {
	addTwoWayConnection(firstNode.node, secondNode.node, firstChan, secondChan)
	firstNode.weights = append(firstNode.weights, weight)
	secondNode.weights = append(secondNode.weights, weight)
}
