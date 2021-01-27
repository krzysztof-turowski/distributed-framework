package lib

import (
	"log"
	"math/rand"
)

func addSynchronizedWeightedConnection(vertices []WeightedGraphNode, i, j, weight int) {
	chans := getSynchronousChannels(2)
	addTwoWayWeightedConnection(
		vertices[i].(*twoWayWeightedGraphNode), vertices[j].(*twoWayWeightedGraphNode),
		chans[0], chans[1],
		weight)
	log.Println("Channel", vertices[i].GetIndex(), "->", vertices[j].GetIndex(), "set up with weight ", weight)
	log.Println("Channel", vertices[j].GetIndex(), "->", vertices[i].GetIndex(), "set up with weight ", weight)
}

func BuildSynchronizedEmptyWeightedGraph(n int, indexGenerator Generator) ([]WeightedGraphNode, Synchronizer) {
	vertices, synchronizer := BuildSynchronizedEmptyGraph(n, indexGenerator)
	weightedVertices := make([]WeightedGraphNode, n)
	for i := range weightedVertices {
		weightedVertices[i] = &twoWayWeightedGraphNode{
			node:    vertices[i].(*twoWayNode),
			weights: make([]int, 0),
		}
	}
	return weightedVertices, synchronizer
}

func BuildSynchronizedWeightedGraphFromAdjencyList(adjacencyList [][][2]int, generator Generator) ([]WeightedGraphNode, Synchronizer) {
	vertices, synchronizer := BuildSynchronizedEmptyWeightedGraph(len(adjacencyList), generator)

	for i, l := range adjacencyList {
		for _, e := range l {
			if j := e[0]; i < j-1 {
				addSynchronizedWeightedConnection(vertices, i, j-1, e[1])
			}
		}
	}

	for _, vertex := range vertices {
		vertex.(*twoWayWeightedGraphNode).shuffleTopology()
	}
	return vertices, synchronizer
}

func BuildSynchronizedRandomConnectedWeightedGraph(n, m, maxWeight int, generator Generator) ([]WeightedGraphNode, Synchronizer) {
	vertices, synchronizer := BuildSynchronizedEmptyWeightedGraph(n, generator)
	possibleEdges := make(map[[2]int]bool)

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			possibleEdges[[2]int{i, j}] = true
		}
	}

	for j := 1; j < n; j++ {
		i := rand.Intn(j)
		delete(possibleEdges, [2]int{i, j})
		addSynchronizedWeightedConnection(vertices, i, j, rand.Intn(maxWeight))
		m--
	}

	leftEdges := make([][2]int, 0)
	for e := range possibleEdges {
		leftEdges = append(leftEdges, e)
	}

	for m > 0 && len(leftEdges) > 0 {
		r := rand.Intn(len(leftEdges))
		e := leftEdges[r]
		addSynchronizedWeightedConnection(vertices, e[0], e[1], rand.Intn(maxWeight))
		l := len(leftEdges)
		leftEdges[r], leftEdges[l-1] = leftEdges[l-1], leftEdges[r]
		leftEdges = leftEdges[:l-1]
		m--
	}

	return vertices, synchronizer
}
