package mst

import (
	"container/heap"
	"fmt"
	"lib"
	"log"
	"reflect"
	"sort"
)

// Priority Queue for sequential MST algorithm

type priorityQueue []*edge

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Empty() bool { return len(pq) == 0 }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].isLess(pq[j])
}

func (pq *priorityQueue) Push(x interface{}) {
	item := x.(*edge)
	*pq = append(*pq, item)
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}

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
	pq       priorityQueue
}

func newSequentialMst(vertices []lib.WeightedGraphNode) *sequentialMst {
	m := &sequentialMst{
		vertices: vertices,
		indexes:  make(map[int]int),
		visited:  make([]bool, len(vertices)),
		pq:       make(priorityQueue, 0),
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

// assuming all edges have unique weights than there is single MST
func verifySynchronizedGHS(vertices []lib.WeightedGraphNode) {
	expected := findMST(vertices)
	for i, v := range vertices {
		r := getTreeEdges(v)
		sort.Ints(r)
		sort.Ints(expected[i])
		if !reflect.DeepEqual(r, expected[i]) {
			panic(fmt.Sprintf("Node %d has tree neighbors %v but should have %v", v.GetIndex(), r, expected[i]))
		}
	}
	log.Printf("Algorithm output is correct\n")
}

func getTreeEdges(v lib.WeightedGraphNode) []int {
	tree := make([]int, 0)
	for i, status := range readState(v).Edges {
		if status == treeEdge {
			tree = append(tree, v.GetOutNeighbors()[i].GetIndex())
		}
	}
	return tree
}
