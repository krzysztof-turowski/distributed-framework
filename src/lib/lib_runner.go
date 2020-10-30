package lib

import (
  "log"
  "math/rand"
  "reflect"
)

type Runner struct {
  n int
  messages int
  inConfirm []chan counterMessage
  outConfirm []chan bool
}

func (r *Runner) Run() {
  cases := make([]reflect.SelectCase, len(r.inConfirm))
  for i, channel := range r.inConfirm {
    cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(channel)}
    r.outConfirm[i] <- true
  }
  finish := 0
  for finish < r.n {
    index, value, _ := reflect.Select(cases)
    message := value.Interface().(*counterMessage)
    if message.finish {
      finish++
      cases[index].Chan = reflect.ValueOf(nil)
    }
    r.messages += message.receivedMessages
    for _, i := range rand.Perm(r.n) {
      log.Println("Node", i, "sent a message")
    }
    r.outConfirm[index] <- true
  }
}

func (r *Runner) GetStats() {
  log.Println("Total messages: ", r.messages)
}
