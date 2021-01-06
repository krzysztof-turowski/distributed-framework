package lib

import (
	"log"
	"math/rand"
)

func BuildSynchronizedEmptyDirectedGraph(n int) ([]Node, Synchronizer) {
	vertices := make([]Node, n)
	inConfirm := make([]chan counterMessage, n)
	outConfirm := make([]chan bool, n)
	for i := range inConfirm {
		inConfirm[i] = make(chan counterMessage)
		outConfirm[i] = make(chan bool)
	}
	rng := GetRandomGenerator()
	for i := range vertices {
		vertices[i] = &oneWayNode{
			index:               rng.Int(),
			size:                n,
			inNeighbors:         make([]<-chan []byte, 0),
			outNeighbors:        make([]chan<- []byte, 0),
			inNeighborsIndices:  make([]int, 0),
			outNeighborsIndices: make([]int, 0),
			stats: statsNode{
				inConfirm:  outConfirm[i],
				outConfirm: inConfirm[i],
			},
		}
		log.Println("Node", vertices[i].GetIndex(), "built")
	}
	return vertices, Synchronizer{n: n, inConfirm: inConfirm, outConfirm: outConfirm}
}

func BuildSynchronizedDirectedRing(n int) ([]Node, Synchronizer) {
	vertices, synchronizer := BuildSynchronizedEmptyDirectedGraph(n)
	chans := getTwoWayChannels(n)
	for i := 0; i < n; i++ {
		addOneWayConnection(
			vertices[i].(*oneWayNode), vertices[(i+1)%n].(*oneWayNode), chans[i])
		log.Println(
			"Channel", vertices[i].GetIndex(), "->", vertices[(i+1)%n].GetIndex(), "set up")
	}
	for _, vertex := range vertices {
		vertex.(*oneWayNode).shuffleTopology()
	}
	return vertices, synchronizer
}

func BuildSynchronizedEmptyGraph(n int, indexGenerator Generator) ([]Node, Synchronizer) {
	vertices := make([]Node, n)
	inConfirm := make([]chan counterMessage, n)
	outConfirm := make([]chan bool, n)
	for i := range inConfirm {
		inConfirm[i] = make(chan counterMessage)
		outConfirm[i] = make(chan bool)
	}
	for i := range vertices {
		vertices[i] = &twoWayNode{
			index:            indexGenerator.Int(),
			size:             n,
			neighbors:        make([]twoWaySynchronousChannel, 0),
			neighborsIndices: make([]int, 0),
			stats: statsNode{
				inConfirm:  outConfirm[i],
				outConfirm: inConfirm[i],
			},
		}
		log.Println("Node", vertices[i].GetIndex(), "built")
	}
	return vertices, Synchronizer{n: n, inConfirm: inConfirm, outConfirm: outConfirm}
}

func BuildSynchronizedRing(n int) ([]Node, Synchronizer) {
	vertices, synchronizer := BuildSynchronizedEmptyGraph(n, GetRandomGenerator())
	chans := getTwoWayChannels(2 * n)
	for i := 0; i < n; i++ {
		addTwoWayConnection(
			vertices[i].(*twoWayNode), vertices[(i+1)%n].(*twoWayNode),
			chans[2*i], chans[(2*i+1)%(2*n)])
		log.Println("Channel", vertices[i].GetIndex(), "->", vertices[(i+1)%n].GetIndex(), "set up")
		log.Println("Channel", vertices[(i+1)%n].GetIndex(), "->", vertices[i].GetIndex(), "set up")
	}
	for _, vertex := range vertices {
		vertex.(*twoWayNode).shuffleTopology()
	}
	return vertices, synchronizer
}

func BuildSynchronizedCompleteGraph(n int) ([]Node, Synchronizer) {
	vertices, synchronizer := BuildSynchronizedEmptyGraph(n, GetRandomGenerator())
	chans := getTwoWayChannels(n * (n - 1))
	counter := 0
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			addTwoWayConnection(
				vertices[i].(*twoWayNode), vertices[j].(*twoWayNode),
				chans[counter], chans[counter+1])
			counter += 2
			log.Println("Channel", vertices[i].GetIndex(), "->", vertices[j].GetIndex(), "set up")
			log.Println("Channel", vertices[j].GetIndex(), "->", vertices[i].GetIndex(), "set up")
		}
	}
	for _, vertex := range vertices {
		vertex.(*twoWayNode).shuffleTopology()
	}
	return vertices, synchronizer
}

func BuildSynchronizedRandomGraph(n int, p float64) ([]Node, Synchronizer) {
	vertices, synchronizer := BuildSynchronizedEmptyGraph(n, GetRandomGenerator())
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if p < rand.Float64() {
				chans := getTwoWayChannels(2)
				addTwoWayConnection(
					vertices[i].(*twoWayNode), vertices[j].(*twoWayNode),
					chans[0], chans[1])
				log.Println("Channel", vertices[i].GetIndex(), "->", vertices[j].GetIndex(), "set up")
				log.Println("Channel", vertices[j].GetIndex(), "->", vertices[i].GetIndex(), "set up")
			}
		}
	}
	for _, vertex := range vertices {
		vertex.(*twoWayNode).shuffleTopology()
	}
	return vertices, synchronizer
}

func BuildSynchronizedRandomTree(n int) ([]Node, Synchronizer) {
	vertices, synchronizer := BuildSynchronizedEmptyGraph(n, GetRandomGenerator())
	chans := getTwoWayChannels(2 * n)
	counter := 0
	for i := 1; i < n; i++ {
		j := rand.Intn(i)
		addTwoWayConnection(
			vertices[i].(*twoWayNode), vertices[j].(*twoWayNode),
			chans[counter], chans[counter+1])
		counter += 2
		log.Println("Channel", vertices[i].GetIndex(), "->", vertices[j].GetIndex(), "set up")
		log.Println("Channel", vertices[j].GetIndex(), "->", vertices[i].GetIndex(), "set up")
	}
	for _, vertex := range vertices {
		vertex.(*twoWayNode).shuffleTopology()
	}
	return vertices, synchronizer
}

func BuildSynchronizedGraphFromAdjacencyList(adjacencyList [][]int) ([]Node, Synchronizer) {
	vertices, synchronizer := BuildSynchronizedEmptyGraph(len(adjacencyList), GetGenerator())
	for i, l := range adjacencyList {
		for _, j := range l {
			if i < j-1 {
				chans := getTwoWayChannels(2)
				addTwoWayConnection(
					vertices[i].(*twoWayNode), vertices[j-1].(*twoWayNode),
					chans[0], chans[1])
				log.Println("Channel", vertices[i].GetIndex(), "->", vertices[j-1].GetIndex(), "set up")
				log.Println("Channel", vertices[j-1].GetIndex(), "->", vertices[i].GetIndex(), "set up")
			}
		}
	}
	for _, vertex := range vertices {
		vertex.(*twoWayNode).shuffleTopology()
	}
	return vertices, synchronizer
}
