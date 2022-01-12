package lib

type Node interface {
	ReceiveAnyMessage() (int, []byte)
	ReceiveMessage(index int) []byte
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

func getAsynchronousChannels(n int) []chan []byte {
	channels := make([]chan []byte, n)
	for i := range channels {
		channels[i] = make(chan []byte)
	}
	return channels
}
