package main

import (
  "flag"
  "io/ioutil"
  "log"
  "testing"
  "leader/directed_ring"
)

var isLogOn = flag.Bool("log", false, "Log output to screen")

func checkLogOutput() {
  if !*isLogOn {
    log.SetOutput(ioutil.Discard)
  }
}

func TestDirectedRingChangRoberts(t *testing.T) {
  checkLogOutput()
  directed_ring.RunChangRoberts(1000)
}

func TestDirectedRingItaiRodeh(t *testing.T) {
  checkLogOutput()
  directed_ring.RunItaiRodeh(1000)
}

func BenchmarkDirectedRingChangRoberts(b *testing.B) {
  log.SetOutput(ioutil.Discard)
  for iteration := 0; iteration < b.N; iteration++ {
    directed_ring.RunChangRoberts(1000)
  }
}

func BenchmarkDirectedRingItaiRodeh(b *testing.B) {
  log.SetOutput(ioutil.Discard)
  for iteration := 0; iteration < b.N; iteration++ {
    directed_ring.RunItaiRodeh(1000)
  }
}
