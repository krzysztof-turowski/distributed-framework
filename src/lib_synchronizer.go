package main

import (
  "log"
  "math/rand"
)

type synchronizer struct {
  n int
  round int
  messages int
  inConfirm []chan counterMessage
  outConfirm []chan bool
}

func (s *synchronizer) synchronize() {
  for s.round = 1; ; s.round++ {
    log.Println("Round", s.round, "started")
    for _, i := range rand.Perm(s.n) {
      s.outConfirm[i] <- true
    }
    finish := false
    for _, i := range rand.Perm(s.n) {
      message := <- s.inConfirm[i]
      finish = (finish || message.finish)
      s.messages += message.sentMessages
      log.Println(
          "Node", i, "sent", message.sentMessages,
          "and received", message.receivedMessages, "messages")
    }
    log.Println("Round", s.round, "finished")
    if finish {
      break
    }
  }
}

func (s *synchronizer) getStats() {
  log.Println("Total messages: ", s.messages)
  log.Println("Total rounds: ", s.round)
}
