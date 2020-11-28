package undirected_graph

import "testing"

func TestUpdateStatus(t *testing.T) {
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
