package main

import (
  "log"
  "math/rand"
)

type counterMessage struct {
  finish bool
  sentMessages int
  receivedMessages int
}

type twoWaySynchronousChannel struct {
  input <-chan []byte
  output chan<- []byte
}

type twoWayNode struct {
  index int
  sentMessages int
  receivedMessages int
  neighbors []twoWaySynchronousChannel
  inConfirm <-chan bool
  outConfirm chan<- counterMessage
}

func getTwoWayChannels(n int) []chan []byte {
  chans := make([]chan []byte, 2 * n)
  for i := range chans {
    chans[i] = make(chan []byte, 1)
  }
  return chans
}

func (v *twoWayNode) receiveMessage(index int) []byte {
  message := <- v.neighbors[index].input
  if message != nil {
    v.receivedMessages++
  }
  return message
}

func (v *twoWayNode) sendMessage(index int, message []byte) {
  v.neighbors[index].output <- message
  if message != nil {
    v.sentMessages++
  }
}

func (v twoWayNode) getChannelsCount() int {
  return len(v.neighbors)
}

func (v twoWayNode) getIndex() int {
  return v.index
}

func (v *twoWayNode) shuffleTopology() {
  rand.Shuffle(len(v.neighbors), func(i, j int) {
      v.neighbors[i], v.neighbors[j] = v.neighbors[j], v.neighbors[i]
  })
}

func (v twoWayNode) run() {
  s := state{}
  for round := 1; ; round++ {
    <- v.inConfirm
    log.Println("Node", v.index, "started")
    finish := process(round, &v, &s)
    log.Println("Node", v.index, "finished")
    v.outConfirm <- counterMessage{
      finish: finish,
      sentMessages: v.sentMessages,
      receivedMessages: v.receivedMessages,
    }
    v.sentMessages, v.receivedMessages = 0, 0
  }
}
