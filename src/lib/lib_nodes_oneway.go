package lib

import (
  "log"
  "math/rand"
)

type oneWayNode struct {
  index int
  state []byte
  inNeighbors []<-chan []byte
  outNeighbors []chan<- []byte
  stats statsNode
}

func getOneWayChannels(n int) []chan []byte {
  channels := make([]chan []byte, n)
  for i := range channels {
    channels[i] = make(chan []byte, 1)
  }
  return channels
}

func (v *oneWayNode) ReceiveMessage(index int) []byte {
  message := <- v.inNeighbors[index]
  if message != nil {
    v.stats.receivedMessages++
  }
  return message
}

func (v *oneWayNode) SendMessage(index int, message []byte) {
  log.Println("Node", v.GetIndex(), "sends message to neighbor", index)
  v.outNeighbors[index] <- message
  if message != nil {
    v.stats.sentMessages++
  }
}

func (v *oneWayNode) GetInChannelsCount() int {
  return len(v.inNeighbors)
}

func (v *oneWayNode) GetOutChannelsCount() int {
  return len(v.outNeighbors)
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

func (v *oneWayNode) StartProcessing() {
  <- v.stats.inConfirm
  log.Println("Node", v.GetIndex(), "started")
}

func (v *oneWayNode) FinishProcessing(finish bool) {
  log.Println("Node", v.GetIndex(), "finished")
  v.stats.outConfirm <- counterMessage{
    finish: finish,
    sentMessages: v.stats.sentMessages,
    receivedMessages: v.stats.receivedMessages,
  }
  v.stats.sentMessages, v.stats.receivedMessages = 0, 0
}

func (v *oneWayNode) shuffleTopology() {
  rand.Shuffle(len(v.inNeighbors), func(i, j int) {
      v.inNeighbors[i], v.inNeighbors[j] = v.inNeighbors[j], v.inNeighbors[i]
  })
  rand.Shuffle(len(v.outNeighbors), func(i, j int) {
      v.outNeighbors[i], v.outNeighbors[j] = v.outNeighbors[j], v.outNeighbors[i]
  })
}

func addOneWayConnection(
    firstNode *oneWayNode, secondNode *oneWayNode, channel chan []byte) {
  firstNode.outNeighbors = append(firstNode.outNeighbors, channel)
  secondNode.inNeighbors = append(secondNode.inNeighbors, channel)
}
