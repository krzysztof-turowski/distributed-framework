package lib

import (
	"fmt"
	"log"
	"math/rand"
)

func addWeightedConnection(vertices []WeightedGraphNode, i, j, weight int) {
	chans := getSynchronousChannels(2)
	addTwoWayWeightedConnection(
		vertices[i].(*twoWayWeightedGraphNode), vertices[j].(*twoWayWeightedGraphNode),
		chans[0], chans[1],
		weight)
	log.Println("Channel", vertices[i].GetIndex(), "->", vertices[j].GetIndex(), "set up with weight ", weight)
	log.Println("Channel", vertices[j].GetIndex(), "->", vertices[i].GetIndex(), "set up with weight ", weight)
}

func BuildEmptyWeightedGraph(n int, indexGenerator Generator) ([]WeightedGraphNode, Runner) {
	vertices, runner := BuildEmptyGraph(n, indexGenerator)
	weightedVertices := make([]WeightedGraphNode, n)
	for i := range weightedVertices {
		weightedVertices[i] = &twoWayWeightedGraphNode{
			node:    vertices[i].(*twoWayNode),
			weights: make([]int, 0),
		}
	}
	return weightedVertices, runner
}

func BuildSynchronizedEmptyWeightedGraph(n int, indexGenerator Generator) ([]WeightedGraphNode, Synchronizer) {
	vertices, runner := BuildEmptyWeightedGraph(n, indexGenerator)
	return vertices, asSynchronizer(runner)
}

func BuildWeightedGraphFromAdjacencyList(adjacencyList [][][2]int, generator Generator) ([]WeightedGraphNode, Runner) {
	vertices, runner := BuildEmptyWeightedGraph(len(adjacencyList), generator)

	for i, l := range adjacencyList {
		for _, e := range l {
			if j := e[0]; i < j-1 {
				addWeightedConnection(vertices, i, j-1, e[1])
			}
		}
	}

	for _, vertex := range vertices {
		vertex.(*twoWayWeightedGraphNode).shuffleTopology()
	}
	return vertices, runner
}

func BuildSynchronizedWeightedGraphFromAdjacencyList(adjacencyList [][][2]int, generator Generator) ([]WeightedGraphNode, Synchronizer) {
	vertices, runner := BuildWeightedGraphFromAdjacencyList(adjacencyList, generator)
	return vertices, asSynchronizer(runner)
}

func buildWeightedRandomTree(n, maxWeight int, generator Generator) ([]WeightedGraphNode, Runner, map[[2]int]bool) {
	vertices, runner := BuildEmptyWeightedGraph(n, generator)
	edgesSet := make(map[[2]int]bool)
	for j := 1; j < n; j++ {
		i := rand.Intn(j)
		addWeightedConnection(vertices, i, j, rand.Intn(maxWeight))
		edgesSet[[2]int{i, j}] = true
		edgesSet[[2]int{j, i}] = true
	}
	return vertices, runner, edgesSet
}

func BuildWeightedRandomTree(n, maxWeight int, generator Generator) ([]WeightedGraphNode, Runner) {
	vertices, runner, _ := buildWeightedRandomTree(n, maxWeight, generator)
	for _, v := range vertices {
		v.(*twoWayWeightedGraphNode).shuffleTopology()
	}
	return vertices, runner
}

func BuildSynchronizedWeightedRandomTree(n, maxWeight int, generator Generator) ([]WeightedGraphNode, Synchronizer) {
	vertices, runner := BuildWeightedRandomTree(n, maxWeight, generator)
	return vertices, asSynchronizer(runner)
}

func BuildRandomConnectedWeightedGraph(n, m, maxWeight int, generator Generator) ([]WeightedGraphNode, Runner) {
	if maxM := n * (n - 1) / 2; m < n-1 || m > maxM {
		panic(fmt.Sprintf("Number of edges should be within range [%d, %d]", n-1, maxM))
	}

	vertices, runner, edgesSet := buildWeightedRandomTree(n, maxWeight, generator)
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
		addWeightedConnection(vertices, e[0], e[1], rand.Intn(maxWeight))
	}
	return vertices, runner
}

func BuildSynchronizedRandomConnectedWeightedGraph(n, m, maxWeight int, generator Generator) ([]WeightedGraphNode, Synchronizer) {
	vertices, runner := BuildRandomConnectedWeightedGraph(n, m, maxWeight, generator)
	return vertices, asSynchronizer(runner)
}