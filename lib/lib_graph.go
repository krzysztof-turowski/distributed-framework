package lib

import (
	"log"
	"math/rand"
)

func asSynchronizer(runner Runner) Synchronizer {
	return Synchronizer{n: runner.n, inConfirm: runner.inConfirm, outConfirm: runner.outConfirm}
}

func BuildEmptyDirectedGraph(n int) ([]Node, Runner) {
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
			index:                rng.Int(),
			size:                 n,
			inNeighborsChannels:  make([]<-chan []byte, 0),
			inNeighbors:          make([]Node, 0),
			outNeighborsChannels: make([]chan<- []byte, 0),
			outNeighbors:         make([]Node, 0),
			stats: statsNode{
				inConfirm:  outConfirm[i],
				outConfirm: inConfirm[i],
			},
		}
		log.Println("Node", vertices[i].GetIndex(), "built")
	}
	return vertices, Runner{n: n, vertices: vertices, inConfirm: inConfirm, outConfirm: outConfirm}
}

func BuildSynchronizedEmptyDirectedGraph(n int) ([]Node, Synchronizer) {
	vertices, runner := BuildEmptyDirectedGraph(n)
	return vertices, asSynchronizer(runner)
}

func BuildDirectedRing(n int) ([]Node, Runner) {
	vertices, runner := BuildEmptyDirectedGraph(n)
	chans := getSynchronousChannels(n)
	for i := 0; i < n; i++ {
		addOneWayConnection(
			vertices[i].(*oneWayNode), vertices[(i+1)%n].(*oneWayNode), chans[i])
		log.Println(
			"Channel", vertices[i].GetIndex(), "->", vertices[(i+1)%n].GetIndex(), "set up")
	}
	for _, vertex := range vertices {
		vertex.(*oneWayNode).shuffleTopology()
	}
	return vertices, runner
}

func BuildSynchronizedDirectedRing(n int) ([]Node, Synchronizer) {
	vertices, runner := BuildDirectedRing(n)
	return vertices, asSynchronizer(runner)
}

func BuildEmptyGraph(n int, indexGenerator Generator) ([]Node, Runner) {
	vertices := make([]Node, n)
	inConfirm := make([]chan counterMessage, n)
	outConfirm := make([]chan bool, n)
	for i := range inConfirm {
		inConfirm[i] = make(chan counterMessage)
		outConfirm[i] = make(chan bool)
	}
	for i := range vertices {
		vertices[i] = &twoWayNode{
			index:             indexGenerator.Int(),
			size:              n,
			neighborsChannels: make([]twoWaySynchronousChannel, 0),
			neighbors:         make([]Node, 0),
			stats: statsNode{
				inConfirm:  outConfirm[i],
				outConfirm: inConfirm[i],
			},
		}
		log.Println("Node", vertices[i].GetIndex(), "built")
	}
	return vertices, Runner{n: n, vertices: vertices, inConfirm: inConfirm, outConfirm: outConfirm}
}

func BuildSynchronizedEmptyGraph(n int, indexGenerator Generator) ([]Node, Synchronizer) {
	vertices, runner := BuildEmptyGraph(n, indexGenerator)
	return vertices, asSynchronizer(runner)
}

func BuildRing(n int) ([]Node, Runner) {
	vertices, runner := BuildEmptyGraph(n, GetRandomGenerator())
	chans := getSynchronousChannels(2 * n)
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
	return vertices, runner
}

func BuildSynchronizedRing(n int) ([]Node, Synchronizer) {
	vertices, runner := BuildRing(n)
	return vertices, asSynchronizer(runner)
}

func BuildCompleteGraph(n int) ([]Node, Runner) {
	vertices, runner := BuildEmptyGraph(n, GetRandomGenerator())
	chans := getSynchronousChannels(n * (n - 1))
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
	return vertices, runner
}

func BuildSynchronizedCompleteGraph(n int) ([]Node, Synchronizer) {
	vertices, runner := BuildCompleteGraph(n)
	return vertices, asSynchronizer(runner)
}

func BuildRandomGraph(n int, p float64) ([]Node, Runner) {
	vertices, runner := BuildEmptyGraph(n, GetRandomGenerator())
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if p < rand.Float64() {
				chans := getSynchronousChannels(2)
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
	return vertices, runner
}

func BuildSynchronizedRandomGraph(n int, p float64) ([]Node, Synchronizer) {
	vertices, runner := BuildRandomGraph(n, p)
	return vertices, asSynchronizer(runner)
}

func BuildRandomTree(n int) ([]Node, Runner) {
	vertices, runner := BuildEmptyGraph(n, GetRandomGenerator())
	chans := getSynchronousChannels(2*n - 2)
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
	return vertices, runner
}

func BuildSynchronizedRandomTree(n int) ([]Node, Synchronizer) {
	vertices, runner := BuildRandomTree(n)
	return vertices, asSynchronizer(runner)
}

func BuildGraphFromAdjacencyList(adjacencyList [][]int, generator Generator) ([]Node, Runner) {
	vertices, runner := BuildEmptyGraph(len(adjacencyList), generator)
	for i, l := range adjacencyList {
		for _, j := range l {
			if i < j-1 {
				chans := getSynchronousChannels(2)
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
	return vertices, runner
}

func BuildSynchronizedGraphFromAdjacencyList(adjacencyList [][]int, generator Generator) ([]Node, Synchronizer) {
	vertices, runner := BuildGraphFromAdjacencyList(adjacencyList, generator)
	return vertices, asSynchronizer(runner)
}

func BuildUndirectedMesh(a int, b int) ([]Node, Runner) {
	log.Println("buliding", a, "x", b, "mesh")
	adjacencyList := make([][]int, a*b)
	for i := range adjacencyList {
		adjacencyList[i] = make([]int, 0)
	}
	for i := 1; i <= a; i++ {
		for j := 1; j <= b; j++ {
			if i-1 > 0 {
				adjacencyList[(i-1)*b+j-1] = append(adjacencyList[(i-1)*b+j-1], (i-2)*b+j)
			}
			if j-1 > 0 {
				adjacencyList[(i-1)*b+j-1] = append(adjacencyList[(i-1)*b+j-1], (i-1)*b+j-1)
			}
			if i+1 <= a {
				adjacencyList[(i-1)*b+j-1] = append(adjacencyList[(i-1)*b+j-1], (i)*b+j)
			}
			if j+1 <= b {
				adjacencyList[(i-1)*b+j-1] = append(adjacencyList[(i-1)*b+j-1], (i-1)*b+j+1)
			}
		}
	}
	return BuildGraphFromAdjacencyList(adjacencyList, GetRandomGenerator())
}

func BuildSynchronizedUndirectedMesh(a int, b int) ([]Node, Synchronizer) {
	vertices, runner := BuildUndirectedMesh(a, b)
	return vertices, asSynchronizer(runner)
}

func BuildHypercube(dim int, oriented bool) ([]Node, Runner) {
	if dim < 0 {
		panic("Dimension cannot be negative")
	}

	log.Println("building", dim, "dimensional cube")
	vertices, runner := BuildEmptyGraph(1<<dim, GetRandomGenerator())
	if len(vertices) == 0 {
		panic("No vertices were created")
	}

	for d := 0; d < dim; d++ {
		for i := range vertices {
			if (i & (1 << d)) > 0 {
				continue
			}

			chans := getSynchronousChannels(2)
			j := i | (1 << d)
			addTwoWayConnection(vertices[i].(*twoWayNode), vertices[j].(*twoWayNode), chans[0], chans[1])
			log.Println("Channel", vertices[i].GetIndex(), "->", vertices[j].GetIndex(), "set up")
			log.Println("Channel", vertices[j].GetIndex(), "->", vertices[i].GetIndex(), "set up")
		}
	}
	if !oriented {
		for _, vertex := range vertices {
			vertex.(*twoWayNode).shuffleTopology()
		}
	}
	rand.Shuffle(len(vertices), func(i, j int) { vertices[i], vertices[j] = vertices[j], vertices[i] })

	return vertices, runner
}

func BuildSynchronizedHypercube(dim int, oriented bool) ([]Node, Synchronizer) {
	vertices, runner := BuildHypercube(dim, oriented)
	return vertices, asSynchronizer(runner)
}

func BuildAsyncEmptyDirectedGraph(n int, indexGenerator Generator) ([]Node, Runner) {
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
			index:                rng.Int(),
			size:                 n,
			inNeighborsChannels:  make([]<-chan []byte, 0),
			inNeighbors:          make([]Node, 0),
			outNeighborsChannels: make([]chan<- []byte, 0),
			outNeighbors:         make([]Node, 0),
			stats: statsNode{
				inConfirm:  outConfirm[i],
				outConfirm: inConfirm[i],
			},
		}
		log.Println("Node", vertices[i].GetIndex(), "built")
	}
	return vertices, Runner{n: n, vertices: vertices, inConfirm: inConfirm, outConfirm: outConfirm}
}

func BuildAsyncCompleteGraphWithLoops(n int) ([]Node, Runner) {
	vertices, runner := BuildAsyncEmptyDirectedGraph(n, GetRandomGenerator())
	chans := getAsynchronousChannels(n * n)
	counter := 0
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			addOneWayConnection(vertices[i].(*oneWayNode), vertices[j].(*oneWayNode), chans[counter])
			counter += 1
			log.Println("Channel", vertices[i].GetIndex(), "->", vertices[j].GetIndex(), "set up")
		}
	}
	for _, vertex := range vertices {
		vertex.(*oneWayNode).shuffleTopology()
	}
	return vertices, runner
}
