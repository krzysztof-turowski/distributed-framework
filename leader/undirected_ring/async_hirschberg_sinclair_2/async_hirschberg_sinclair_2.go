package async_hirschberg_sinclair_2

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/krzysztof-turowski/distributed-framework/lib"
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

type message struct {
	MessageType MessageType
	Value       int
	Num         uint64
	Maxnum      uint64
}

type state struct {
	MyStatus ElectionStatus
}

func initialize(v lib.Node) bool {
	s := state{MyStatus: CandidateStatus}
	sEncoded, _ := json.Marshal(s)
	v.SetState(sEncoded)

	data := message{From, v.GetIndex(), 0, 1}
	sendBoth(v, data)
	return true
}

func process(v lib.Node) bool {
	maxnum := uint64(1)
	var jsonReceived []byte
	var chanIndex int
	var heardL bool
	var heardR bool
	var receivedData message
	counter := 0
	final := false
	var s state
	for {
		chanIndex, jsonReceived = v.ReceiveAnyMessage()
		_ = json.Unmarshal(jsonReceived, &receivedData)
		_ = json.Unmarshal(v.GetState(), &s)

		switch receivedData.MessageType {
		case From:
			receivedMessageIsFrom(v, receivedData, chanIndex)
			_ = json.Unmarshal(v.GetState(), &s)
			if s.MyStatus == WonStatus {
				counter++
				if counter == 2 {
					dataToSend := message{End, v.GetIndex(), 0, maxnum}
					sendPass(v, &dataToSend, chanIndex)
				}
			}

		case No, Ok:
			receivedMessageIsOkOrNo(v, receivedData, chanIndex)
			_ = json.Unmarshal(v.GetState(), &s)
			if receivedData.MessageType == No {
				s.MyStatus = LostStatus
			}
			if receivedData.Value == v.GetIndex() {
				if chanIndex == 0 {
					heardL = true
				} else {
					heardR = true
				}
				if heardL && heardR && s.MyStatus == CandidateStatus {
					heardL = false
					heardR = false
					maxnum *= 2
					dataToSend := initSendingData(From, v.GetIndex(), 0, maxnum)
					sendBoth(v, dataToSend)
				}
			}
		case End:
			if s.MyStatus == LostStatus {
				sendPass(v, &receivedData, chanIndex)
			}
			final = true

		default:
			//nothing to do
		}

		sEncoded, _ := json.Marshal(s)
		v.SetState(sEncoded)

		if final {
			break
		}
	}

	if s.MyStatus == WonStatus {
		log.Println("The Leader has index ", v.GetIndex())
	}
	return true
}

func run(v lib.Node, wg *sync.WaitGroup) {
	defer wg.Done()
	v.StartProcessing()
	isFinished := initialize(v)

	isFinished = process(v)
	v.FinishProcessing(isFinished)

}

func check(vertices []lib.Node) {
	var leader lib.Node
	var s state
	for _, v := range vertices {
		_ = json.Unmarshal(v.GetState(), &s)
		if s.MyStatus == WonStatus {
			leader = v
			break
		}
	}
	if leader == nil {
		panic("There is no Leader on the undirected ring")
	}
	for _, v := range vertices {
		_ = json.Unmarshal(v.GetState(), &s)
		if v != leader {
			if s.MyStatus == WonStatus {
				panic(fmt.Sprint(
					"There is more then one Leader on the undirected ring"))
			}
			if s.MyStatus == CandidateStatus {
				panic(fmt.Sprint("Node ", v.GetIndex(), " has state ", CandidateStatus))
			}
		} else {
			if s.MyStatus != WonStatus {
				panic("There is more then one Leader on the undirected ring")
			}
		}
	}
}

func Run(n int) (int, int) {
	vertices, runner := lib.BuildRing(n)
	var wg sync.WaitGroup
	wg.Add(n)
	for _, v := range vertices {
		log.Println("Node", v.GetIndex(), "about to run")
		go run(v, &wg)
	}
	runner.Run(true)
	check(vertices)
	wg.Wait()
	return runner.GetStats()
}

func sendBoth(v lib.Node, dataToSend message) {
	jsonToSend, _ := json.Marshal(dataToSend)
	v.SendMessage(0, jsonToSend)
	v.SendMessage(1, jsonToSend)
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func getChanToPrev(chanIndex int) int {
	return chanIndex
}

func getChanToNext(chanIndex int) int {
	return abs(chanIndex - 1)
}

func sendEcho(v lib.Node, dataToSend *message, chanIndex int) {
	channel := getChanToPrev(chanIndex)
	jsonToSend, _ := json.Marshal(dataToSend)
	v.SendMessage(channel, jsonToSend)
}

func sendPass(v lib.Node, dataToSend *message, chanIndex int) {
	channel := getChanToNext(chanIndex)
	jsonToSend, _ := json.Marshal(dataToSend)
	v.SendMessage(channel, jsonToSend)
}

func receivedMessageIsFrom(v lib.Node, receivedData message, chanIndex int) {
	if receivedData.Value < v.GetIndex() {
		dataToSend := initSendingData(No, receivedData.Value, 0, 0)
		sendEcho(v, &dataToSend, chanIndex)
		return
	}
	var s state
	_ = json.Unmarshal(v.GetState(), &s)
	if receivedData.Value > v.GetIndex() {
		s.MyStatus = LostStatus
		sEncoded, _ := json.Marshal(s)
		v.SetState(sEncoded)
		if receivedData.Num+1 < receivedData.Maxnum {
			dataToSend := receivedData
			dataToSend.Num += 1
			sendPass(v, &dataToSend, chanIndex)
		} else {
			dataToSend := initSendingData(Ok, receivedData.Value, 0, 0)
			sendEcho(v, &dataToSend, chanIndex)
		}
		return
	}

	//value == v.GetIndex()
	s.MyStatus = WonStatus
	sEncoded, _ := json.Marshal(s)
	v.SetState(sEncoded)
}

func receivedMessageIsOkOrNo(v lib.Node, receivedData message, chanIndex int) {
	if receivedData.Value != v.GetIndex() {
		sendPass(v, &receivedData, chanIndex)
	} else {
		//nice to see you back, my message!
	}
}

func initSendingData(messageType MessageType, value int, num uint64, maxnum uint64) message {
	return message{
		MessageType: messageType,
		Value:       value,
		Num:         num,
		Maxnum:      maxnum,
	}
}