package main

import (
  "encoding/json"
  "lib"
  "log"
  "os"
  "strconv"
  "time"
)

func initialize(v lib.Node, _ interface{}) bool {
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

func process(v lib.Node, _ interface{}, round int) bool {
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
  return round == 3
}

func run(v lib.Node) {
  var s int
  v.StartProcessing()
  finish := initialize(v, s)
  data, _ := json.Marshal(s)
  v.SetState(data)
  v.FinishProcessing(finish)
  
  for round := 1; ; round++ {
    v.StartProcessing()
    json.Unmarshal(v.GetState(), &s)
    finish := process(v, s, round)
    data, _ := json.Marshal(s)
    v.SetState(data)
    v.FinishProcessing(finish)
  }
}

func main() {
  n, _ := strconv.Atoi(os.Args[len(os.Args) - 1])
  vertices, synchronizer := lib.BuildSynchronizedCompleteGraph(n)
  for _, v := range vertices {
    log.Println("Node", v.GetIndex(), "about to run")
    go run(v)
  }
  synchronizer.Synchronize()
  time.Sleep(5 * time.Millisecond)
  synchronizer.GetStats()
}
