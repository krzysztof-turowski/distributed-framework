package main

import (
  "log"
  "os"
  "strconv"
  "time"
)

func main() {
  n, _ := strconv.Atoi(os.Args[len(os.Args) - 1])
  vertices, synchronizer := buildSynchronizedCompleteGraph(n)
  for _, v := range vertices {
    log.Println("Node", v.index, "about to run")
    go v.run()
  }
  synchronizer.synchronize()
  time.Sleep(5 * time.Millisecond)
  synchronizer.getStats()
}
