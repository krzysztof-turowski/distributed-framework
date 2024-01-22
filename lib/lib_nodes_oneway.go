package lib

import (
	"log"
	"math/rand"
	"reflect"
)

type oneWayNode struct {
	index, size          int
	state                []byte
	inNeighborsChannels  []<-chan []byte
	outNeighborsChannels []chan<- []byte
	inNeighbors          []Node
	outNeighbors         []Node
	inNeighborsCases     []reflect.SelectCase
	stats                statsNode
}

func (v *oneWayNode) ReceiveMessage(index int) []byte {
	message := <-v.inNeighborsChannels[index]
	if message != nil {
		v.stats.receivedMessages++
	}
	return message
}

func (v *oneWayNode) ReceiveAnyMessage() (int, []byte) {
	for {
		index, value, ok := reflect.Select(v.inNeighborsCases)
		if !ok {
			continue
		}
		message := value.Interface().([]byte)
		if message != nil {
			v.stats.receivedMessages++
		}
		return index, message
	}
}

func (v *oneWayNode) ReceiveMessageIfAvailable(index int) []byte {
	neighborsCasesDefault := make([]reflect.SelectCase, 2)
	neighborsCasesDefault[0] = v.inNeighborsCases[index]
	neighborsCasesDefault[1] = reflect.SelectCase{Dir: reflect.SelectDefault}
	_, value, ok := reflect.Select(neighborsCasesDefault)
	if !ok {
		return nil
	}
	message := value.Interface().([]byte)
	if message != nil {
		v.stats.receivedMessages++
	}
	return message
}

func (v *oneWayNode) SendMessage(index int, message []byte) {
	log.Println("Node", v.GetIndex(), "sends message to neighbor", index)
	v.outNeighborsChannels[index] <- message
	if message != nil {
		v.stats.sentMessages++
	}
}

func (v *oneWayNode) GetInChannelsCount() int {
	return len(v.inNeighborsChannels)
}

func (v *oneWayNode) GetOutChannelsCount() int {
	return len(v.outNeighborsChannels)
}

func (v *oneWayNode) GetInNeighbors() []Node {
	return v.inNeighbors
}

func (v *oneWayNode) GetOutNeighbors() []Node {
	return v.outNeighbors
}

func (v *oneWayNode) GetIndex() int {
	return v.index
}

func (v *oneWayNode) GetState() []byte {
	return v.state
}

func (v *oneWayNode) SetState(state []byte) {
	v.state = state
}

func (v *oneWayNode) GetSize() int {
	return v.size
}

func (v *oneWayNode) StartProcessing() {
	<-v.stats.inConfirm
	log.Println("Node", v.GetIndex(), "started")
}

func (v *oneWayNode) FinishProcessing(finish bool) {
	log.Println("Node", v.GetIndex(), "finished")
	v.stats.outConfirm <- counterMessage{
		finish:           finish,
		sentMessages:     v.stats.sentMessages,
		receivedMessages: v.stats.receivedMessages,
	}
	v.stats.sentMessages, v.stats.receivedMessages = 0, 0
}

func (v *oneWayNode) IgnoreFutureMessages() {
	go func() {
		for {
			_, _, _ = reflect.Select(v.inNeighborsCases)
		}
	}()
}

func (v *oneWayNode) Close() {
	for _, channel := range v.outNeighborsChannels {
		close(channel)
	}
	for _, channel := range v.inNeighborsChannels {
		for range channel {
		}
	}
}

func (v *oneWayNode) shuffleTopology() {
	rand.Shuffle(len(v.inNeighborsChannels), func(i, j int) {
		v.inNeighborsChannels[i], v.inNeighborsChannels[j] = v.inNeighborsChannels[j], v.inNeighborsChannels[i]
		v.inNeighbors[i], v.inNeighbors[j] = v.inNeighbors[j], v.inNeighbors[i]
		v.inNeighborsCases[i], v.inNeighborsCases[j] = v.inNeighborsCases[j], v.inNeighborsCases[i]
		v.outNeighborsChannels[i], v.outNeighborsChannels[j] = v.outNeighborsChannels[j], v.outNeighborsChannels[i]
		v.outNeighbors[i], v.outNeighbors[j] = v.outNeighbors[j], v.outNeighbors[i]
	})
}

func addOneWayConnection(firstNode, secondNode *oneWayNode, channel chan []byte) {
	firstNode.outNeighborsChannels = append(firstNode.outNeighborsChannels, channel)
	firstNode.outNeighbors = append(firstNode.outNeighbors, secondNode)
	secondNode.inNeighborsChannels = append(secondNode.inNeighborsChannels, channel)
	secondNode.inNeighbors = append(secondNode.inNeighbors, firstNode)
	secondNode.inNeighborsCases = append(
		secondNode.inNeighborsCases,
		reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(channel)})
}

func addTwoWayConnection(firstNode, secondNode *oneWayNode, firstChan, secondChan chan []byte) {
	addOneWayConnection(firstNode, secondNode, firstChan)
	addOneWayConnection(secondNode, firstNode, secondChan)
}
