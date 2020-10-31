package lib

import (
  "log"
  "math/rand"
  "time"
)

type Synchronizer struct {
  n int
  round int
  messages int
  inConfirm []chan counterMessage
  outConfirm []chan bool
}

func (s *Synchronizer) Synchronize(interval time.Duration) {
  for finish := 0; finish < s.n; s.round++ {
    log.Println("Round", s.round, "initialized")
    for _, i := range rand.Perm(s.n) {
      if s.outConfirm[i] != nil {
        s.outConfirm[i] <- true
      }
    }
    log.Println("Round", s.round, "started")
    for _, i := range rand.Perm(s.n) {
      if s.inConfirm[i] == nil {
        continue
      }
      message := <- s.inConfirm[i]
      if message.finish {
        finish++
        s.inConfirm[i], s.outConfirm[i] = nil, nil
      }
      s.messages += message.receivedMessages
      log.Println(
          "Node", i, "sent", message.sentMessages,
          "and received", message.receivedMessages, "messages")
    }
    log.Println("Round", s.round, "finished")
    time.Sleep(interval)
  }
}

func (s *Synchronizer) GetStats() {
  log.Println("Total messages: ", s.messages)
  log.Println("Total rounds: ", s.round)
}
