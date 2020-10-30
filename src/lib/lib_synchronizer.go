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
  finish := false
  for s.round = 0; !finish; s.round++ {
    log.Println("Round", s.round, "initialized")
    for _, i := range rand.Perm(s.n) {
      s.outConfirm[i] <- true
    }
    log.Println("Round", s.round, "started")
    finish = false
    for _, i := range rand.Perm(s.n) {
      message := <- s.inConfirm[i]
      finish = (finish || message.finish)
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
