package mst

import (
	"container/heap"
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

// helper functions for Adjency List type

type adjencyList [][]int

func (al adjencyList) addEdge(vertices []lib.WeightedGraphNode, i, j int) {
	al[i] = append(al[i], vertices[j].GetIndex())
	al[j] = append(al[j], vertices[i].GetIndex())
}

func newAdjencyList(n int) adjencyList {
	al := make(adjencyList, n)
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

func findMST(vertices []lib.WeightedGraphNode) adjencyList {
	al := newAdjencyList(len(vertices))
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

// assuming all edges have unique edges than there is single MST
func verifySynchSHS(vertices []lib.WeightedGraphNode) bool {
	expected := findMST(vertices)
	correct := true
	for i, v := range vertices {
		r := getTreeEdges(readState(v).Edges, v)
		sort.Ints(r)
		sort.Ints(expected[i])
		if !reflect.DeepEqual(r, expected[i]) {
			log.Printf("Node %d has tree neighbors %v but should have %v\n", v.GetIndex(), r, expected[i])
			correct = false
		}
	}
	if correct {
		log.Printf("Algorithm output is correct\n")
	} else {
		log.Printf("Algorithm output is wrong\n")
	}
	return correct
}

func getTreeEdges(edges []edgeStatus, v lib.WeightedGraphNode) []int {
	tree := make([]int, 0)
	for i, status := range edges {
		if status == treeEdge {
			tree = append(tree, v.GetOutNeighbors()[i].GetIndex())
		}
	}
	return tree
}
