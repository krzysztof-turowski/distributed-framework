package directed_ring

import (
  "encoding/json"
  "lib"
  "log"
  "time"
)

type modeType string

const (
  unknown modeType = "unknown"
  nonleader = "nonleader"
  leader = "leader"
)

type state struct {
  Min int
  Status modeType
}

type message struct {
  Min int
  Mode modeType
}

func send(v lib.Node, s state, mode modeType) {
  outMessage, _ := json.Marshal(message{Min: s.Min, Mode: mode})
  v.SendMessage(0, outMessage)
  data, _ := json.Marshal(s)
  v.SetState(data)
}

func receive(v lib.Node) (state, message) {
  var s state
  json.Unmarshal(v.GetState(), &s)
  inMessage := v.ReceiveMessage(0)
  var m message
  json.Unmarshal(inMessage, &m)
  return s, m
}

func initialize(v lib.Node) bool {
  s := state{Min: v.GetIndex(), Status: unknown}
  send(v, s, unknown)
  return false
}

func process(v lib.Node, round int) bool {
  s, m := receive(v)
  if s.Status == unknown {
    if m.Mode == leader {
      s.Status = nonleader
      send(v, s, leader)
    } else if m.Min == v.GetIndex() {
      s.Status = leader
      send(v, s, leader)
    } else if m.Min < v.GetIndex() {
      s.Min = m.Min
      send(v, s, unknown)
    } else {
      send(v, s, unknown)
    }
  }
  return (s.Status == leader && m.Mode == leader) || s.Status == nonleader
}

func run(v lib.Node) {
  v.StartProcessing()
  finish := initialize(v)
  v.FinishProcessing(finish)
  
  for round := 1; !finish; round++ {
    v.StartProcessing()
    finish = process(v, round)
    v.FinishProcessing(finish)
  }
}

func RunChangRoberts(n int) {
  vertices, synchronizer := lib.BuildSynchronizedDirectedRing(n)
  for _, v := range vertices {
    log.Println("Node", v.GetIndex(), "about to run")
    go run(v)
  }
  synchronizer.Synchronize(10 * time.Millisecond)
  synchronizer.GetStats()
}
