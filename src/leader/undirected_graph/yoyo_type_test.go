package undirected_graph

import "testing"

func TestUpdateStatus(t *testing.T) {
	checkLogOutput()

	s := newStateYoYo(3)

	s.EdgeStates[0] = out
	s.EdgeStates[1] = out
	s.EdgeStates[2] = out
	s.UpdateStatus()
	if s.Status != source {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be source, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = pruned
	s.EdgeStates[1] = out
	s.EdgeStates[2] = out
	s.UpdateStatus()
	if s.Status != source {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be source, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = out
	s.EdgeStates[1] = pruned
	s.EdgeStates[2] = out
	s.UpdateStatus()
	if s.Status != source {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be source, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = out
	s.EdgeStates[1] = out
	s.EdgeStates[2] = pruned
	s.UpdateStatus()
	if s.Status != source {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be source, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = pruned
	s.EdgeStates[1] = pruned
	s.EdgeStates[2] = out
	s.UpdateStatus()
	if s.Status != source {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be source, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = pruned
	s.EdgeStates[1] = out
	s.EdgeStates[2] = pruned
	s.UpdateStatus()
	if s.Status != source {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be source, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = out
	s.EdgeStates[1] = pruned
	s.EdgeStates[2] = pruned
	s.UpdateStatus()
	if s.Status != source {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be source, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = in
	s.EdgeStates[1] = in
	s.EdgeStates[2] = in
	s.UpdateStatus()
	if s.Status != sink {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be sink, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = pruned
	s.EdgeStates[1] = in
	s.EdgeStates[2] = in
	s.UpdateStatus()
	if s.Status != sink {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be sink, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = in
	s.EdgeStates[1] = pruned
	s.EdgeStates[2] = in
	s.UpdateStatus()
	if s.Status != sink {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be sink, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = in
	s.EdgeStates[1] = in
	s.EdgeStates[2] = pruned
	s.UpdateStatus()
	if s.Status != sink {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be sink, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = pruned
	s.EdgeStates[1] = pruned
	s.EdgeStates[2] = in
	s.UpdateStatus()
	if s.Status != sink {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be sink, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = pruned
	s.EdgeStates[1] = in
	s.EdgeStates[2] = pruned
	s.UpdateStatus()
	if s.Status != sink {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be sink, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = in
	s.EdgeStates[1] = pruned
	s.EdgeStates[2] = pruned
	s.UpdateStatus()
	if s.Status != sink {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be sink, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = pruned
	s.EdgeStates[1] = pruned
	s.EdgeStates[2] = pruned
	s.UpdateStatus()
	if s.Status != detached {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be detached, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = in
	s.EdgeStates[1] = in
	s.EdgeStates[2] = out
	s.UpdateStatus()
	if s.Status != internal {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be internal, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = in
	s.EdgeStates[1] = out
	s.EdgeStates[2] = in
	s.UpdateStatus()
	if s.Status != internal {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be internal, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = in
	s.EdgeStates[1] = out
	s.EdgeStates[2] = out
	s.UpdateStatus()
	if s.Status != internal {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be internal, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = in
	s.EdgeStates[1] = out
	s.EdgeStates[2] = pruned
	s.UpdateStatus()
	if s.Status != internal {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be internal, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = in
	s.EdgeStates[1] = pruned
	s.EdgeStates[2] = out
	s.UpdateStatus()
	if s.Status != internal {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be internal, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = out
	s.EdgeStates[1] = in
	s.EdgeStates[2] = in
	s.UpdateStatus()
	if s.Status != internal {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be internal, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = out
	s.EdgeStates[1] = in
	s.EdgeStates[2] = out
	s.UpdateStatus()
	if s.Status != internal {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be internal, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = out
	s.EdgeStates[1] = in
	s.EdgeStates[2] = pruned
	s.UpdateStatus()
	if s.Status != internal {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be internal, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = out
	s.EdgeStates[1] = out
	s.EdgeStates[2] = in
	s.UpdateStatus()
	if s.Status != internal {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be internal, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = out
	s.EdgeStates[1] = pruned
	s.EdgeStates[2] = in
	s.UpdateStatus()
	if s.Status != internal {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be internal, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = pruned
	s.EdgeStates[1] = in
	s.EdgeStates[2] = out
	s.UpdateStatus()
	if s.Status != internal {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be internal, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}

	s.EdgeStates[0] = pruned
	s.EdgeStates[1] = out
	s.EdgeStates[2] = in
	s.UpdateStatus()
	if s.Status != internal {
		t.Errorf("Status of vertex with edges (%s,%s,%s) should be internal, but is %s",
			s.EdgeStates[0], s.EdgeStates[1], s.EdgeStates[2], s.Status)
	}
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

	for i := 0; i < 6; i++ {
		s.EdgeStates[i] = in
	}
	s.EdgeStates[6] = out
	s.LastMessages[0] = 5
	s.LastMessages[1] = 4
	s.LastMessages[2] = 3
	s.LastMessages[3] = 5
	s.LastMessages[4] = 4
	s.LastMessages[5] = 5
	s.LastMessages[6] = 4

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
