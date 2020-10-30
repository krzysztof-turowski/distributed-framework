package lib

import (
  "log"
  "math/rand"
)

func BuildSynchronizedEmptyGraph(n int) ([]Node, Synchronizer) {
  vertices := make([]Node, n)
  inConfirm := make([]chan counterMessage, n)
  outConfirm := make([]chan bool, n)
  for i := range inConfirm {
    inConfirm[i] = make(chan counterMessage)
    outConfirm[i] = make(chan bool)
  }
  for i := range vertices {
    vertices[i] = &twoWayNode{
      index: i,
      neighbors: make([]twoWaySynchronousChannel, 0),
      stats: statsNode{
        inConfirm: outConfirm[i],
        outConfirm: inConfirm[i],
      },
    }
    log.Println("Node", vertices[i].GetIndex(), "built")
  }
  return vertices, Synchronizer{n: n, inConfirm: inConfirm, outConfirm: outConfirm,}
}

func BuildSynchronizedRing(n int) ([]Node, Synchronizer) {
  vertices, synchronizer := BuildSynchronizedEmptyGraph(n)
  chans := getTwoWayChannels(2 * n)
  for i := 0; i < n; i++ {
    addTwoWayConnection(
        vertices[i].(*twoWayNode), vertices[(i + 1) % n].(*twoWayNode),
        chans[2 * i], chans[(2 * i + 1) % (2 * n)])
    log.Println("Channel", i, "->", (i + 1) % n, "set up")
    log.Println("Channel", (i + 1) % n, "->", i, "set up")
  }
  for _, vertex := range vertices {
    vertex.(*twoWayNode).shuffleTopology()
  }
  return vertices, synchronizer
}

func BuildSynchronizedCompleteGraph(n int) ([]Node, Synchronizer) {
  vertices, synchronizer := BuildSynchronizedEmptyGraph(n)
  chans := getTwoWayChannels(n * (n - 1))
  counter := 0
  for i := 0; i < n; i++ {
    for j := i + 1; j < n; j++ {
      addTwoWayConnection(
          vertices[i].(*twoWayNode), vertices[j].(*twoWayNode),
          chans[counter], chans[counter + 1])
      counter += 2
      log.Println("Channel", i, "->", j, "set up")
      log.Println("Channel", j, "->", i, "set up")
    }
  }
  for _, vertex := range vertices {
    vertex.(*twoWayNode).shuffleTopology()
  }
  return vertices, synchronizer
}

func BuildSynchronizedRandomGraph(n int, p float64) ([]Node, Synchronizer) {
  vertices, synchronizer := BuildSynchronizedEmptyGraph(n)
  for i := 0; i < n; i++ {
    for j := i + 1; j < n; j++ {
      if p < rand.Float64() {
        chans := getTwoWayChannels(2)
        addTwoWayConnection(
            vertices[i].(*twoWayNode), vertices[j].(*twoWayNode),
            chans[0], chans[1])
        log.Println("Channel", i, "->", j, "set up")
        log.Println("Channel", j, "->", i, "set up")
      }
    }
  }
  for _, vertex := range vertices {
    vertex.(*twoWayNode).shuffleTopology()
  }
  return vertices, synchronizer
}

func BuildSynchronizedRandomTree(n int) ([]Node, Synchronizer) {
  vertices, synchronizer := BuildSynchronizedEmptyGraph(n)
  chans := getTwoWayChannels(2 * n)
  counter := 0
  for i := 1; i < n; i++ {
    j := rand.Intn(i)
    addTwoWayConnection(
        vertices[i].(*twoWayNode), vertices[j].(*twoWayNode),
        chans[counter], chans[counter + 1])
    counter += 2
    log.Println("Channel", i, "->", j, "set up")
    log.Println("Channel", j, "->", i, "set up")
  }
  for _, vertex := range vertices {
    vertex.(*twoWayNode).shuffleTopology()
  }
  return vertices, synchronizer
}
