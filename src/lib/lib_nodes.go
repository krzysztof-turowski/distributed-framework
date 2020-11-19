package lib

type Node interface {
  ReceiveMessage(index int) []byte
  SendMessage(index int, message []byte)
  GetInChannelsCount() int
  GetOutChannelsCount() int
  GetIndex() int
  GetState() []byte
  SetState(state []byte)
  GetSize() int
  StartProcessing()
  FinishProcessing(finish bool)
}

type counterMessage struct {
  finish bool
  sentMessages int
  receivedMessages int
}

type statsNode struct {
  sentMessages int
  receivedMessages int
  inConfirm <-chan bool
  outConfirm chan<- counterMessage
}
