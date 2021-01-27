package mst

/*  EDGES
 * Synchronized GHS algorithm for finding Minimal Spanning Tree requires every edge to have a unique weight.
 * To achieve that for every u,v in E we define new edge weight as a triple (w, v, u) where w is the original
 * weight and v < u. We order such triplets lexicographically.
 */
type edge struct {
	Weight   int
	SmallerV int
	LargerV  int
}

func newEdge(weight, u, v int) *edge {
	if u > v {
		u, v = v, u
	}
	return &edge{Weight: weight, SmallerV: u, LargerV: v}
}

func (e *edge) isLess(f *edge) bool {
	if e.Weight != f.Weight {
		return e.Weight < f.Weight
	}
	if e.SmallerV != f.SmallerV {
		return e.SmallerV < f.SmallerV
	}
	return e.LargerV < f.LargerV
}

func (e *edge) isConnected(v int) bool {
	return v == e.SmallerV || v == e.LargerV
}

type edgeStatus int

const (
	outgoingEdge edgeStatus = iota
	treeEdge
	rejectedEdge
)
