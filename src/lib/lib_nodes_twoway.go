package lib

import (
	"log"
	"math/rand"
)

type twoWaySynchronousChannel struct {
	input  <-chan []byte
	output chan<- []byte
}

type twoWayNode struct {
	index, size       int
	state             []byte
	neighborsChannels []twoWaySynchronousChannel
	neighbors         []Node
	stats             statsNode
}

func (v *twoWayNode) ReceiveMessage(index int) []byte {
	message := <-v.neighborsChannels[index].input
	if message != nil {
		v.stats.receivedMessages++
	}
	return message
}

func (v *twoWayNode) SendMessage(index int, message []byte) {
	log.Println("Node", v.GetIndex(), "sends message to neighbor", index)
	v.neighborsChannels[index].output <- message
	if message != nil {
		v.stats.sentMessages++
	}
}

func (v *twoWayNode) GetInChannelsCount() int {
	return len(v.neighborsChannels)
}

func (v *twoWayNode) GetOutChannelsCount() int {
	return len(v.neighborsChannels)
}

func (v *twoWayNode) GetInNeighbors() []Node {
	return v.neighbors
}

func (v *twoWayNode) GetOutNeighbors() []Node {
	return v.neighbors
}

func (v *twoWayNode) GetIndex() int {
	return v.index
}

func (v *twoWayNode) GetState() []byte {
	return v.state
}

func (v *twoWayNode) SetState(state []byte) {
	v.state = state
}

func (v *twoWayNode) GetSize() int {
	return v.size
}

func (v *twoWayNode) StartProcessing() {
	<-v.stats.inConfirm
	log.Println("Node", v.GetIndex(), "started")
}

func (v *twoWayNode) FinishProcessing(finish bool) {
	log.Println("Node", v.GetIndex(), "finished")
	v.stats.outConfirm <- counterMessage{
		finish:           finish,
		sentMessages:     v.stats.sentMessages,
		receivedMessages: v.stats.receivedMessages,
	}
	v.stats.sentMessages, v.stats.receivedMessages = 0, 0
}

func (v *twoWayNode) shuffleTopology() {
	rand.Shuffle(len(v.neighborsChannels), func(i, j int) {
		v.neighborsChannels[i], v.neighborsChannels[j] = v.neighborsChannels[j], v.neighborsChannels[i]
		v.neighbors[i], v.neighbors[j] = v.neighbors[j], v.neighbors[i]
	})
}

func addTwoWayConnection(
	firstNode *twoWayNode, secondNode *twoWayNode,
	firstChan chan []byte, secondChan chan []byte) {
	firstNode.neighborsChannels = append(
		firstNode.neighborsChannels,
		twoWaySynchronousChannel{input: secondChan, output: firstChan})
	firstNode.neighbors = append(firstNode.neighbors, secondNode)
	secondNode.neighborsChannels = append(
		secondNode.neighborsChannels,
		twoWaySynchronousChannel{input: firstChan, output: secondChan})
	secondNode.neighbors = append(secondNode.neighbors, firstNode)
}
