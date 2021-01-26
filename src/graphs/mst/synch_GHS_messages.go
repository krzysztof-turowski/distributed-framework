package mst

import (
	"encoding/json"
	"lib"
)

type messageType int

const (
	nilMessage      messageType = 0
	msgVerifyEdge   messageType = 1
	msgProposeEdge  messageType = 2
	msgChooseEdge   messageType = 3
	msgConnectEdge  messageType = 4
	msgElectNewRoot messageType = 5
	msgCompleated   messageType = 6
)

type messageSynchGHS struct {
	Type  messageType
	MWOE  *edge // MWOE - Minimal Weight Outgoing Edge
	Index int
	Root  int
}

func sendMessageSynchGHS(v lib.WeightedGraphNode, index int, message *messageSynchGHS) {
	outMessage, _ := json.Marshal(message)
	v.SendMessage(index, outMessage)
}

func receiveMessageSynchGHS(v lib.WeightedGraphNode, index int) *messageSynchGHS {
	var inMessage messageSynchGHS
	if data := v.ReceiveMessage(index); data != nil {
		if err := json.Unmarshal(data, &inMessage); err == nil {
			return &inMessage
		}
	}
	return nil
}
