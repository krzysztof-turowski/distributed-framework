package sync_hirschberg_sinclair_second_edition

import (
	"encoding/json"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type ElectionStatus string
type MessageType string

const (
	CandidateStatus ElectionStatus = "candidate"
	LostStatus      ElectionStatus = "lost"
	WonStatus       ElectionStatus = "won"
)
const (
	No   MessageType = "no"
	Ok   MessageType = "ok"
	From MessageType = "from"
	End  MessageType = "end"
)

type ICandidate interface {
	SelectLeader() int
}

type SendingData struct {
	MessageType MessageType
	Value       int
	Num         uint64
	Maxnum      uint64
}

type Candidate struct {
	myValue     int
	myStatus    ElectionStatus
	leftInput   <-chan []byte
	leftOutput  chan<- []byte
	rightInput  <-chan []byte
	rightOutput chan<- []byte
}

func (cand *Candidate) Sendboth(dataToSend SendingData) {
	jsonToSend, _ := json.Marshal(dataToSend)
	cand.sendTo(cand.leftOutput, jsonToSend)
	cand.sendTo(cand.rightOutput, jsonToSend)
}

func (cand *Candidate) GetChanToPrev(replyFromL bool) chan<- []byte {
	if replyFromL {
		return cand.leftOutput
	}
	return cand.rightOutput
}

func (cand *Candidate) GetChanToNext(replyFromL bool) chan<- []byte {
	if replyFromL {
		return cand.rightOutput
	}
	return cand.leftOutput
}

func (cand *Candidate) Sendecho(dataToSend *SendingData, replyFromL bool) {
	channel := cand.GetChanToPrev(replyFromL)
	jsonToSend, _ := json.Marshal(dataToSend)
	cand.sendTo(channel, jsonToSend)
}

func (cand *Candidate) Sendpass(dataToSend *SendingData, replyFromL bool) {
	channel := cand.GetChanToNext(replyFromL)
	jsonToSend, _ := json.Marshal(dataToSend)
	cand.sendTo(channel, jsonToSend)
}

func (cand *Candidate) ReceivedMessageIsFrom(receivedData SendingData, replyFromL bool) {
	if receivedData.Value < cand.myValue {
		dataToSend := cand.InitSendingData(No, receivedData.Value, 0, 0)
		cand.Sendecho(&dataToSend, replyFromL)
		return
	}
	if receivedData.Value > cand.myValue {
		cand.myStatus = LostStatus
		if receivedData.Num+1 < receivedData.Maxnum {
			dataToSend := receivedData
			dataToSend.Num += 1
			cand.Sendpass(&dataToSend, replyFromL)
		} else {
			dataToSend := cand.InitSendingData(Ok, receivedData.Value, 0, 0)
			cand.Sendecho(&dataToSend, replyFromL)
		}
		return
	}
	//value == cand.myValue
	cand.myStatus = WonStatus
}

func (cand *Candidate) ReceivedMessageIsOkOrNo(receivedData SendingData, replyFromL bool) {
	if receivedData.Value != cand.myValue {
		cand.Sendpass(&receivedData, replyFromL)
	} else {
		//nice to see you back, my message!
	}
}

func (cand *Candidate) SelectLeader() int {
	maxnum := uint64(1)
	var jsonReceived []byte
	var replyFromL bool
	var LReplyed bool
	var RReplyed bool
	var receivedData SendingData

	dataToSend := cand.InitSendingData(From, cand.myValue, 0, maxnum)
	cand.Sendboth(dataToSend)
	counter := 0
	final := false
	for {
		select {
		case jsonReceived = <-cand.leftInput:
			replyFromL = true
		case jsonReceived = <-cand.rightInput:
			replyFromL = false
		}
		_ = json.Unmarshal(jsonReceived, &receivedData)
		switch receivedData.MessageType {
		case From:
			cand.ReceivedMessageIsFrom(receivedData, replyFromL)
			if cand.myStatus == WonStatus {
				counter++
				if counter == 2 {
					dataToSend.MessageType = End
					cand.Sendpass(&dataToSend, replyFromL)
				}
			}

		case No, Ok:
			cand.ReceivedMessageIsOkOrNo(receivedData, replyFromL)
			if receivedData.MessageType == No {
				cand.myStatus = LostStatus
			}
			if receivedData.Value == cand.myValue {
				if replyFromL {
					LReplyed = true
				} else {
					RReplyed = true
				}
				if LReplyed && RReplyed && cand.myStatus == CandidateStatus {
					LReplyed = false
					RReplyed = false
					maxnum *= 2
					dataToSend := cand.InitSendingData(From, cand.myValue, 0, maxnum)
					cand.Sendboth(dataToSend)
				}
			}
		case End:
			if cand.myStatus == LostStatus {
				cand.Sendpass(&receivedData, replyFromL)
			}
			final = true

		default:
			//nothing to do
		}

		if final {
			break
		}
	}

	if cand.myStatus == WonStatus {
		return cand.myValue
	} else {
		return -1
	}
}

func NewCandidate(
	myValue int,
	leftInput <-chan []byte,
	leftOutput chan<- []byte,
	rightInput <-chan []byte,
	rightOutput chan<- []byte) ICandidate {
	return &Candidate{
		myValue:     myValue,
		myStatus:    CandidateStatus,
		leftInput:   leftInput,
		leftOutput:  leftOutput,
		rightInput:  rightInput,
		rightOutput: rightOutput,
	}
}

func (cand *Candidate) sendTo(channel chan<- []byte, jsonToSend []byte) {
	go func() { channel <- jsonToSend }()
}

func (cand *Candidate) InitSendingData(messageType MessageType, value int, num uint64, maxnum uint64) SendingData {
	return SendingData{
		MessageType: messageType,
		Value:       value,
		Num:         num,
		Maxnum:      maxnum,
	}
}

func BuildUndirectedRingAndRun(N int, argsArr []int) {
	firstInput1 := make(chan []byte)
	firstOutput1 := make(chan []byte)

	prevOutput := firstInput1
	prevInput := firstOutput1

	writeChannel := make(chan int)
	var wg sync.WaitGroup
	wg.Add(N)
	runner := func(candidate ICandidate) {
		defer wg.Done()
		leader := candidate.SelectLeader()
		writeChannel <- leader
	}

	//arr := generateShuffledArray(N)
	var arrBits []bool

	for it := 0; it < N; it++ {
		nextOutput := make(chan []byte)
		nextInput := make(chan []byte)

		var input1, input2 <-chan []byte
		var output1, output2 chan<- []byte
		input1 = prevOutput
		output1 = prevInput
		input2 = nextInput
		output2 = nextOutput
		if it == N-1 {
			input1 = prevOutput
			output1 = prevInput
			output2 = firstInput1
			input2 = firstOutput1
		}

		orient := generateRandomBit()
		if orient {
			input1 = nextInput
			output1 = nextOutput
			input2 = prevOutput
			output2 = prevInput
			if it == N-1 {
				input2 = prevOutput
				output2 = prevInput
				output1 = firstInput1
				input1 = firstOutput1
			}
		}
		arrBits = append(arrBits, orient)

		go runner(NewCandidate(argsArr[it], input1, output1, input2, output2))
		prevOutput = nextOutput
		prevInput = nextInput

	}
	realLeader := -1
	for it := 0; it < N; it++ {
		leader := <-writeChannel
		if realLeader < leader {
			realLeader = leader
		}
	}
	log.Println(realLeader)
	wg.Wait()
}

func GenerateShuffledArray(N int) []int {
	runtime.GOMAXPROCS(1)
	arr := make([]int, N)
	for i := 0; i < N; i++ {
		arr[i] = i
	}

	// Shuffle the array
	//rand.Seed(time.Now().UnixNano())
	rand.Shuffle(N, func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
	return arr
}

type randBit struct {
	R *rand.Rand
}

func generateRandomBit() bool {
	rB := randBit{
		R: rand.New(rand.NewSource(time.Now().UTC().UnixNano())),
	}
	return rB.R.Intn(2) == 0
}
