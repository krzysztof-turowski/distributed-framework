package main

import (
  "encoding/json"
  "log"
)

type state struct {
}

type node interface {
  receiveMessage(index int) []byte
  sendMessage(index int, message []byte)
  getChannelsCount() int
  getIndex() int
}

func process(round int, v node, s *state) bool {
  outMessageA, _ := json.Marshal(round)
  outMessageB := []byte(nil)
  if round == 1 {
    for i := 0; i < v.getChannelsCount() / 2; i++ {
      v.sendMessage(i, outMessageA)
    }
    for i := v.getChannelsCount() / 2; i < v.getChannelsCount(); i++ {
      v.sendMessage(i, outMessageB)
    }
  } else {
    receivedMessages := [][]byte(nil)
    for i := 0; i < v.getChannelsCount(); i++ {
      receivedMessages = append(receivedMessages, v.receiveMessage(i))
    }
    log.Println("Node", v.getIndex(), "received", receivedMessages)
    for i := 0; i < v.getChannelsCount() / 2; i++ {
      v.sendMessage(i, outMessageA)
    }
    for i := v.getChannelsCount() / 2; i < v.getChannelsCount(); i++ {
      v.sendMessage(i, outMessageB)
    }
  }
  return round == 3 && v.getIndex() == 3
}
