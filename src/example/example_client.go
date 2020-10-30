package main

import (
  "encoding/json"
  "lib"
  "log"
)

type state struct {
}

func initialize(v lib.Node, s state) bool {
  outMessageA, _ := json.Marshal(0)
  outMessageB := []byte(nil)
  for i := 0; i < v.GetOutChannelsCount() / 2; i++ {
    v.SendMessage(i, outMessageA)
  }
  for i := v.GetOutChannelsCount() / 2; i < v.GetOutChannelsCount(); i++ {
    v.SendMessage(i, outMessageB)
  }
  return false
}

func process(v lib.Node, s state, round int) bool {
  outMessageA, _ := json.Marshal(round)
  outMessageB := []byte(nil)
  receivedMessages := [][]byte(nil)
  for i := 0; i < v.GetInChannelsCount(); i++ {
    receivedMessages = append(receivedMessages, v.ReceiveMessage(i))
  }
  log.Println("Node", v.GetIndex(), "received", receivedMessages)
  for i := 0; i < v.GetOutChannelsCount() / 2; i++ {
    v.SendMessage(i, outMessageA)
  }
  for i := v.GetOutChannelsCount() / 2; i < v.GetOutChannelsCount(); i++ {
    v.SendMessage(i, outMessageB)
  }
  return round == 3 && v.GetIndex() == 3
}
