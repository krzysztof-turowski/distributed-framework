package mst

import (
	"container/heap"
	"github.com/krzysztof-turowski/distributed-framework/lib"
)

// helper functions for Adjacency List type

type adjacencyList [][]int

func (al adjacencyList) addEdge(vertices []lib.WeightedGraphNode, i, j int) {
	al[i] = append(al[i], vertices[j].GetIndex())
	al[j] = append(al[j], vertices[i].GetIndex())
}

func newAdjacencyList(n int) adjacencyList {
	al := make(adjacencyList, n)
	for i := range al {
		al[i] = make([]int, 0)
	}
	return al
}

// Sequential algorithm for finding MST

type sequentialMst struct {
	vertices []lib.WeightedGraphNode
	indexes  map[int]int
	visited  []bool
	pq       lib.PriorityQueue
}

func newSequentialMst(vertices []lib.WeightedGraphNode) *sequentialMst {
	m := &sequentialMst{
		vertices: vertices,
		indexes:  make(map[int]int),
		visited:  make([]bool, len(vertices)),
		pq:       make(lib.PriorityQueue, 0),
	}
	heap.Init(&m.pq)
	for i, v := range vertices {
		m.indexes[v.GetIndex()] = i
		m.visited[i] = false
	}
	return m
}

func (m *sequentialMst) visit(index int) {
	v := m.vertices[index]
	m.visited[index] = true
	for i, u := range v.GetOutNeighbors() {
		if !m.visited[m.indexes[u.GetIndex()]] {
			heap.Push(&m.pq, newEdge(v.GetOutWeights()[i], v.GetIndex(), u.GetIndex()))
		}
	}
}

func findMST(vertices []lib.WeightedGraphNode) adjacencyList {
	al := newAdjacencyList(len(vertices))
	if len(vertices) == 0 {
		return al
	}

	m := newSequentialMst(vertices)
	m.visit(0)
	for !m.pq.Empty() {
		e := heap.Pop(&m.pq).(*edge)
		if i := m.indexes[e.LargerV]; !m.visited[i] {
			al.addEdge(vertices, i, m.indexes[e.SmallerV])
			m.visit(i)
		}
		if i := m.indexes[e.SmallerV]; !m.visited[i] {
			al.addEdge(vertices, i, m.indexes[e.LargerV])
			m.visit(i)
		}
	}
	return al
}
