package lib

import (
	"log"
	"reflect"
)

type Runner struct {
	n          int
	messages   int
	inConfirm  []chan counterMessage
	outConfirm []chan bool
}

func (r *Runner) Run() {
	cases := make([]reflect.SelectCase, len(r.inConfirm))
	for i, channel := range r.inConfirm {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(channel)}
		r.outConfirm[i] <- true
	}
	for finish := 0; finish < r.n; {
		index, value, _ := reflect.Select(cases)
		message := value.Interface().(*counterMessage)
		if message.finish {
			finish++
			cases[index].Chan, r.outConfirm[index] = reflect.ValueOf(nil), nil
		}
		r.messages += message.receivedMessages
		log.Println(
			"Node number", index+1, "received", message.receivedMessages,
			"and sent", message.sentMessages, "messages")
		if r.outConfirm[index] != nil {
			r.outConfirm[index] <- true
		}
	}
}

func (r *Runner) GetStats() {
	log.Println("Total messages: ", r.messages)
}
