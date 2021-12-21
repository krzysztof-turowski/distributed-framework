package lib

import (
	"fmt"
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

func BuildSynchronizedWeightedGraphFromAdjacencyList(adjacencyList [][][2]int, generator Generator) ([]WeightedGraphNode, Synchronizer) {
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

func buildSynchronizedWeightedRandomTree(n, maxWeight int, generator Generator) ([]WeightedGraphNode, Synchronizer, map[[2]int]bool) {
	vertices, synchronizer := BuildSynchronizedEmptyWeightedGraph(n, generator)
	edgesSet := make(map[[2]int]bool)
	for j := 1; j < n; j++ {
		i := rand.Intn(j)
		addSynchronizedWeightedConnection(vertices, i, j, rand.Intn(maxWeight))
		edgesSet[[2]int{i, j}] = true
		edgesSet[[2]int{j, i}] = true
	}
	return vertices, synchronizer, edgesSet
}

func BuildSynchronizedWeightedRandomTree(n, maxWeight int, generator Generator) ([]WeightedGraphNode, Synchronizer) {
	vertices, synchronizer, _ := buildSynchronizedWeightedRandomTree(n, maxWeight, generator)
	for _, v := range vertices {
		v.(*twoWayWeightedGraphNode).shuffleTopology()
	}
	return vertices, synchronizer
}

func BuildSynchronizedRandomConnectedWeightedGraph(n, m, maxWeight int, generator Generator) ([]WeightedGraphNode, Synchronizer) {
	if maxM := n * (n - 1) / 2; m < n-1 || m > maxM {
		panic(fmt.Sprintf("Number of edges should be within range [%d, %d]", n-1, maxM))
	}

	vertices, synchronizer, edgesSet := buildSynchronizedWeightedRandomTree(n, maxWeight, generator)
	possibleEdges := make([][2]int, 0)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if _, usedEdge := edgesSet[[2]int{i, j}]; !usedEdge {
				possibleEdges = append(possibleEdges, [2]int{i, j})
			}
		}
	}
	rand.Shuffle(len(possibleEdges), func(i, j int) {
		possibleEdges[i], possibleEdges[j] = possibleEdges[j], possibleEdges[i]
	})
	m = m - n + 1
	for _, e := range possibleEdges[:m] {
		addSynchronizedWeightedConnection(vertices, e[0], e[1], rand.Intn(maxWeight))
	}
	return vertices, synchronizer
}
