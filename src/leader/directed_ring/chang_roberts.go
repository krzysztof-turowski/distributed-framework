package directed_ring

import (
  "encoding/json"
  "fmt"
  "lib"
  "log"
)

type modeType string

const (
  pass modeType = "pass"
  unknown = "unknown"
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
  var m message
  if inMessage := v.ReceiveMessageOrNil(0); inMessage != nil {
    json.Unmarshal(inMessage, &m)
  }
  return s, m
}

func initializeChangRoberts(v lib.Node) bool {
  s := state{Min: v.GetIndex(), Status: unknown}
  send(v, s, unknown)
  return false
}

func processChangRoberts(v lib.Node, round int) bool {
  if s, m := receive(v); s.Status == unknown {
    if m.Mode == leader {
      s.Status = nonleader
      send(v, s, leader)
      return true
    } else if m.Mode == unknown {
      if m.Min == v.GetIndex() {
        s.Status = leader
        send(v, s, leader)
      } else if m.Min < s.Min {
        s.Min = m.Min
        send(v, s, unknown)
      }
    }
  } else if s.Status == leader {
    if m.Mode == leader {
      return true
    }
  }
  return false
}

func runChangRoberts(v lib.Node) {
  v.StartProcessing()
  finish := initializeChangRoberts(v)
  v.FinishProcessing(finish)
  
  for round := 1; !finish; round++ {
    v.StartProcessing()
    finish = processChangRoberts(v, round)
    v.FinishProcessing(finish)
  }
}

func checkChangRoberts(vertices []lib.Node) {
  var lead_node lib.Node
  var s state
  for _, v := range vertices {
    json.Unmarshal(v.GetState(), &s)
    if s.Status == leader {
      lead_node = v
      break
    }
  }
  if lead_node == nil {
    panic("There is no leader on the directed ring")
  }
  for _, v := range vertices {
    json.Unmarshal(v.GetState(), &s)
    if v == lead_node {
      if v.GetIndex() != s.Min {
        panic(fmt.Sprintf("Leader has index", v.GetIndex(), "but minimum", s.Min))
      }
    } else {
      if s.Status == leader {
        panic(fmt.Sprintf(
            "Multiple leaders on the directed ring:", lead_node.GetIndex(), v.GetIndex()))
      }
      if s.Status != nonleader {
        panic(fmt.Sprintf("Node", v.GetIndex(), "has state", s.Status))
      }
      if lead_node.GetIndex() != s.Min {
        panic(fmt.Sprintf(
            "Leader has index", lead_node.GetIndex(), "but node", v.GetIndex(),
            "has minimum", s.Min))
      }
    }
  }
}

func RunChangRoberts(n int) {
  vertices, synchronizer := lib.BuildSynchronizedDirectedRing(n)
  for _, v := range vertices {
    log.Println("Node", v.GetIndex(), "about to run")
    go runChangRoberts(v)
  }
  synchronizer.Synchronize(0)
  synchronizer.GetStats()
  checkChangRoberts(vertices)
}
