package undirected_graph

import (
	"encoding/json"
	"flag"
	"github.com/krzysztof-turowski/distributed-framework/lib"
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

func (v *testNode) ReceiveAnyMessage() (int, []byte) {
	return -1, nil
}

func (v *testNode) SendMessage(index int, message []byte) {

}

func (v *testNode) GetInChannelsCount() int {
	return 1
}

func (v *testNode) GetOutChannelsCount() int {
	return 1
}

func (v *testNode) GetInNeighbors() []lib.Node {
	return nil
}

func (v *testNode) GetOutNeighbors() []lib.Node {
	return nil
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

func testUpdateStatus(t *testing.T, edge0, edge1, edge2 edgeStateType, expectedStatus statusType) {
	s := newStateYoYo(3)
	s.EdgeStates[0] = edge0
	s.EdgeStates[1] = edge1
	s.EdgeStates[2] = edge2

	s.UpdateStatus()
	if s.Status != expectedStatus {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be %s, but is %s",
			edge0, edge1, edge2, expectedStatus, s.Status)
	}
}

func TestUpdateStatus(t *testing.T) {
	checkLogOutput()

	testUpdateStatus(t, out, out, out, source)
	testUpdateStatus(t, pruned, out, out, source)
	testUpdateStatus(t, out, pruned, out, source)
	testUpdateStatus(t, out, out, pruned, source)
	testUpdateStatus(t, pruned, pruned, out, source)
	testUpdateStatus(t, pruned, out, pruned, source)
	testUpdateStatus(t, out, pruned, pruned, source)

	testUpdateStatus(t, in, in, in, sink)
	testUpdateStatus(t, pruned, in, in, sink)
	testUpdateStatus(t, in, pruned, in, sink)
	testUpdateStatus(t, in, in, pruned, sink)
	testUpdateStatus(t, pruned, pruned, in, sink)
	testUpdateStatus(t, pruned, in, pruned, sink)
	testUpdateStatus(t, in, pruned, pruned, sink)

	testUpdateStatus(t, pruned, pruned, pruned, detached)

	testUpdateStatus(t, in, in, out, internal)
	testUpdateStatus(t, in, out, in, internal)
	testUpdateStatus(t, in, out, out, internal)
	testUpdateStatus(t, in, out, pruned, internal)
	testUpdateStatus(t, in, pruned, out, internal)
	testUpdateStatus(t, out, in, in, internal)
	testUpdateStatus(t, out, in, out, internal)
	testUpdateStatus(t, out, in, pruned, internal)
	testUpdateStatus(t, out, out, in, internal)
	testUpdateStatus(t, out, pruned, in, internal)
	testUpdateStatus(t, pruned, in, out, internal)
	testUpdateStatus(t, pruned, out, in, internal)
}

func TestFlipEdge(t *testing.T) {
	checkLogOutput()

	s := newStateYoYo(2)

	s.EdgeStates[1] = in
	s.FlipEdge(1)
	if s.EdgeStates[1] != out {
		t.Errorf("Status of edge[1] should be out, but is %s", s.EdgeStates[1])
	}

	s.EdgeStates[1] = out
	s.FlipEdge(1)
	if s.EdgeStates[1] != in {
		t.Errorf("Status of edge[1] should be in, but is %s", s.EdgeStates[1])
	}
}

func TestPreprocessPruning_pruneDefault(t *testing.T) {
	checkLogOutput()

	s := newStateYoYo(2)

	s.EdgeStates[0] = in
	s.EdgeStates[1] = in
	s.LastMessages[0] = 5
	s.LastMessages[1] = 5

	_, pruneDefault := s.PreprocessPruning(0)
	if !pruneDefault {
		t.Errorf("pruneDefault should be true, since vertex is sink and received the same values, but is false")
	}

	s.LastMessages[1] = 3
	_, pruneDefault = s.PreprocessPruning(0)
	if pruneDefault {
		t.Errorf("pruneDefault should be false, since vertex received 2 different values, but is true")
	}

	s.EdgeStates[1] = out
	s.LastMessages[1] = 5
	_, pruneDefault = s.PreprocessPruning(1)
	if pruneDefault {
		t.Errorf("pruneDefault should be false, since vertex is not sink, but is true")
	}
}

func TestPreprocessPruning_requestPrune(t *testing.T) {
	checkLogOutput()

	s := newStateYoYo(7)
	s.EdgeStates = []edgeStateType{in, in, in, in, in, in, out}
	s.LastMessages = []int{5, 4, 3, 5, 4, 5, 4}

	requestPrune, _ := s.PreprocessPruning(1)
	req := make(map[int][]int)
	req[5] = make([]int, 2)
	req[4] = make([]int, 2)
	req[3] = make([]int, 2)
	for i := 0; i < 6; i++ {
		if requestPrune[i] {
			req[s.LastMessages[i]][1]++
		} else {
			req[s.LastMessages[i]][0]++
		}
	}

	if req[5][0] != 1 {
		t.Errorf("Number of %s prune requests for edges with value %d should be %d, but is %d",
			"negative", 5, 1, req[5][0])
	}
	if req[5][1] != 2 {
		t.Errorf("Number of %s prune requests for edges with value %d should be %d, but is %d",
			"positive", 5, 2, req[5][1])
	}
	if req[4][0] != 1 {
		t.Errorf("Number of %s prune requests for edges with value %d should be %d, but is %d",
			"negative", 4, 1, req[4][0])
	}
	if req[4][1] != 1 {
		t.Errorf("Number of %s prune requests for edges with value %d should be %d, but is %d",
			"positive", 4, 1, req[4][1])
	}
	if req[3][0] != 1 {
		t.Errorf("Number of %s prune requests for edges with value %d should be %d, but is %d",
			"negative", 3, 1, req[3][0])
	}
	if req[3][1] != 0 {
		t.Errorf("Number of %s prune requests for edges with value %d should be %d, but is %d",
			"positive", 3, 0, req[3][1])
	}
}

func TestRunSynchronizedGHSSantoroExample(t *testing.T) {
	// Graph from Santoro, Design and Analysis, sec. 3.8.3
	adjacencyList := [][]int{
		{7, 9, 10},
		{3, 8, 11},
		{2, 4, 15},
		{3, 16},
		{7, 8, 12},
		{10, 13},
		{1, 5, 9},
		{2, 5, 12},
		{1, 7},
		{1, 6, 13},
		{2, 15, 16},
		{5, 8, 14, 17},
		{6, 10},
		{12, 18, 19},
		{3, 11, 16},
		{4, 11, 15},
		{12, 18, 19},
		{14, 17},
		{14, 17},
	}
	RunYoYo(lib.BuildSynchronizedGraphFromAdjacencyList(adjacencyList, lib.GetGenerator()))
}
