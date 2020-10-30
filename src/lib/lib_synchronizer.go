package lib

import (
  "log"
  "math/rand"
)

type Synchronizer struct {
  n int
  round int
  messages int
  inConfirm []chan counterMessage
  outConfirm []chan bool
}

func (s *Synchronizer) Synchronize() {
  finish := 0
  for s.round = 0; finish < s.n; s.round++ {
    log.Println("Round", s.round, "initialized")
    for _, i := range rand.Perm(s.n) {
      s.outConfirm[i] <- true
    }
    log.Println("Round", s.round, "started")
    for _, i := range rand.Perm(s.n) {
      message := <- s.inConfirm[i]
      if message.finish {
        finish++
      }
      s.messages += message.receivedMessages
      log.Println(
          "Node", i, "sent", message.sentMessages,
          "and received", message.receivedMessages, "messages")
    }
    log.Println("Round", s.round, "finished")
  }
}

func (s *Synchronizer) GetStats() {
  log.Println("Total messages: ", s.messages)
  log.Println("Total rounds: ", s.round)
}
