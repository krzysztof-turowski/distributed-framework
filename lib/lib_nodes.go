package lib

import "time"

type Node interface {
	ReceiveMessage(index int) []byte
	ReceiveAnyMessage() (int, []byte)
	ReceiveMessageIfAvailable(index int) []byte
	ReceiveMessageWithTimeout(index int, timeout time.Duration) []byte
	SendMessage(index int, message []byte)
	GetInChannelsCount() int
	GetOutChannelsCount() int
	GetInNeighbors() []Node
	GetOutNeighbors() []Node
	GetIndex() int
	GetState() []byte
	SetState(state []byte)
	GetSize() int
	StartProcessing()
	FinishProcessing(finish bool)
	IgnoreFutureMessages()
	Close() // only for use by Runner
}

type counterMessage struct {
	finish           bool
	sentMessages     int
	receivedMessages int
}

type statsNode struct {
	sentMessages     int
	receivedMessages int
	inConfirm        <-chan bool
	outConfirm       chan<- counterMessage
}

func getSynchronousChannels(n int) []chan []byte {
	channels := make([]chan []byte, n)
	for i := range channels {
		channels[i] = make(chan []byte, 1)
	}
	return channels
}
