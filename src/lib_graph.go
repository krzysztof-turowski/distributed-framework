package main

import (
  "log"
  "math/rand"
)

func addTwoWayConnection(
    firstNode *twoWayNode, secondNode *twoWayNode, firstChan chan []byte, secondChan chan []byte) {
  firstNode.neighbors = append(
    firstNode.neighbors,
    twoWaySynchronousChannel{ input: secondChan, output: firstChan, })
  secondNode.neighbors = append(
    secondNode.neighbors,
    twoWaySynchronousChannel{ input: firstChan, output: secondChan, })
}

func buildSynchronizedEmptyGraph(n int) ([]twoWayNode, synchronizer) {
  vertices := make([]twoWayNode, n)
  inConfirm := make([]chan counterMessage, n)
  outConfirm := make([]chan bool, n)
  for i := range inConfirm {
    inConfirm[i] = make(chan counterMessage)
    outConfirm[i] = make(chan bool)
  }
  for i := range vertices {
    vertices[i] = twoWayNode{
      index: i,
      neighbors: make([]twoWaySynchronousChannel, 0),
      inConfirm: outConfirm[i],
      outConfirm: inConfirm[i],
    }
    log.Println("Node", vertices[i].index, "built")
  }
  return vertices, synchronizer{n: n, inConfirm: inConfirm, outConfirm: outConfirm,}
}

func buildSynchronizedRing(n int) ([]twoWayNode, synchronizer) {
  vertices, synchronizer := buildSynchronizedEmptyGraph(n)
  chans := getTwoWayChannels(2 * n)
  for i := 0; i < n; i++ {
    addTwoWayConnection(
        &vertices[i], &vertices[(i + 1) % n], chans[2 * i], chans[(2 * i + 1) % (2 * n)])
    log.Println("Channel", i, "->", (i + 1) % n, "set up")
    log.Println("Channel", (i + 1) % n, "->", i, "set up")
  }
  for _, vertex := range vertices {
    vertex.shuffleTopology()
  }
  return vertices, synchronizer
}

func buildSynchronizedCompleteGraph(n int) ([]twoWayNode, synchronizer) {
  vertices, synchronizer := buildSynchronizedEmptyGraph(n)
  chans := getTwoWayChannels(n * (n - 1))
  counter := 0
  for i := 0; i < n; i++ {
    for j := i + 1; j < n; j++ {
      addTwoWayConnection(&vertices[i], &vertices[j], chans[counter], chans[counter + 1])
      counter += 2
      log.Println("Channel", i, "->", j, "set up")
      log.Println("Channel", j, "->", i, "set up")
    }
  }
  for _, vertex := range vertices {
    vertex.shuffleTopology()
  }
  return vertices, synchronizer
}

func buildSynchronizedRandomGraph(n int, p float64) ([]twoWayNode, synchronizer) {
  vertices, synchronizer := buildSynchronizedEmptyGraph(n)
  for i := 0; i < n; i++ {
    for j := i + 1; j < n; j++ {
      if p < rand.Float64() {
        chans := getTwoWayChannels(2)
        addTwoWayConnection(&vertices[i], &vertices[j], chans[0], chans[1])
        log.Println("Channel", i, "->", j, "set up")
        log.Println("Channel", j, "->", i, "set up")
      }
    }
  }
  for _, vertex := range vertices {
    vertex.shuffleTopology()
  }
  return vertices, synchronizer
}

func buildSynchronizedRandomTree(n int) ([]twoWayNode, synchronizer) {
  vertices, synchronizer := buildSynchronizedEmptyGraph(n)
  chans := getTwoWayChannels(2 * n)
  counter := 0
  for i := 1; i < n; i++ {
    j := rand.Intn(i)
    addTwoWayConnection(&vertices[i], &vertices[j], chans[counter], chans[counter + 1])
    counter += 2
    log.Println("Channel", i, "->", j, "set up")
    log.Println("Channel", j, "->", i, "set up")
  }
  for _, vertex := range vertices {
    vertex.shuffleTopology()
  }
  return vertices, synchronizer
}
