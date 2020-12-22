package undirected_graph

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"testing"
)

var isLogOn = flag.Bool("log", false, "Log output to screen")

func checkLogOutput() {
	if !*isLogOn {
		log.SetOutput(ioutil.Discard)
	}
}

type testNode struct {
	index  int
	toRecv int
	state  []byte
}

func (v *testNode) ReceiveMessage(index int) []byte {
	bytes, _ := json.Marshal(messageYoYo{Content: v.toRecv})
	return bytes
}

func (v *testNode) SendMessage(index int, message []byte) {

}

func (v *testNode) GetInChannelsCount() int {
	return 1
}

func (v *testNode) GetOutChannelsCount() int {
	return 1
}

func (v *testNode) GetIndex() int {
	return v.index
}

func (v *testNode) GetState() []byte {
	return v.state
}

func (v *testNode) SetState(state []byte) {
	v.state = state
}

func (v *testNode) GetSize() int {
	return 2
}

func (v *testNode) StartProcessing() {
}

func (v *testNode) FinishProcessing(finish bool) {
}

func TestYoYoRandom(t *testing.T) {
	checkLogOutput()
	RunYoYoRandom(1000, 0.25)
}

func TestProcessSetup(t *testing.T) {
	checkLogOutput()

	v1 := &testNode{index: 1, toRecv: 2}
	v2 := &testNode{index: 2, toRecv: 1}
	s1 := newStateYoYo(1)
	s2 := newStateYoYo(1)
	setStateYoYo(v1, s1)
	setStateYoYo(v2, s2)

	processSetup(v1, s1)
	processSetup(v2, s2)

	if s1.EdgeStates[0] != out {
		t.Errorf("First node's only edge should be oriented %s, but is %s", out, s1.EdgeStates[0])
	}
	if s2.EdgeStates[0] != in {
		t.Errorf("Second node's only edge should be oriented %s, but is %s", in, s2.EdgeStates[0])
	}
	if s1.Status != source {
		t.Errorf("First node status should be %v, but is %v", source, s1.Status)
	}
	if s2.Status != sink {
		t.Errorf("Second node status should be %v, but is %v", sink, s2.Status)
	}
}
