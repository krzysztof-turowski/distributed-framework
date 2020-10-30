package main

import (
  "encoding/json"
  "lib"
  "log"
  "os"
  "strconv"
  "time"
)

func run(v lib.Node) {
  var s state
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
