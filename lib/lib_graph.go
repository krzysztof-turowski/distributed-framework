package lib

import (
	"log"
	"math"
	"math/rand"
)

func asSynchronizer(runner Runner) Synchronizer {
	return Synchronizer{n: runner.n, inConfirm: runner.inConfirm, outConfirm: runner.outConfirm}
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
		vertices[i] = &oneWayNode{
			index:                indexGenerator.Int(),
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

func BuildSynchronizedEmptyGraph(n int, indexGenerator Generator) ([]Node, Synchronizer) {
	vertices, runner := BuildEmptyGraph(n, indexGenerator)
	return vertices, asSynchronizer(runner)
}

func BuildDirectedRing(n int) ([]Node, Runner) {
	vertices, runner := BuildEmptyGraph(n, GetRandomGenerator())
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

func BuildRing(n int) ([]Node, Runner) {
	vertices, runner := BuildEmptyGraph(n, GetRandomGenerator())
	chans := getSynchronousChannels(2 * n)
	for i := 0; i < n; i++ {
		addTwoWayConnection(
			vertices[i].(*oneWayNode), vertices[(i+1)%n].(*oneWayNode),
			chans[2*i], chans[(2*i+1)%(2*n)])
		log.Println("Channel", vertices[i].GetIndex(), "->", vertices[(i+1)%n].GetIndex(), "set up")
		log.Println("Channel", vertices[(i+1)%n].GetIndex(), "->", vertices[i].GetIndex(), "set up")
	}
	for _, vertex := range vertices {
		vertex.(*oneWayNode).shuffleTopology()
	}
	return vertices, runner
}

func BuildRingWithOriginalTopology(n int) ([]Node, Runner) {
	vertices, runner := BuildEmptyGraph(n, GetRandomGenerator())
	chans := getSynchronousChannels(2 * n)
	for i := 0; i < n; i++ {
		addTwoWayConnection(
			vertices[i].(*oneWayNode), vertices[(i+1)%n].(*oneWayNode),
			chans[2*i], chans[(2*i+1)%(2*n)])
		log.Println("Channel", vertices[i].GetIndex(), "->", vertices[(i+1)%n].GetIndex(), "set up")
		log.Println("Channel", vertices[(i+1)%n].GetIndex(), "->", vertices[i].GetIndex(), "set up")
	}
	return vertices, runner
}

func BuildSynchronizedRingWithOriginalTopology(n int) ([]Node, Synchronizer) {
	vertices, runner := BuildRingWithOriginalTopology(n)
	return vertices, asSynchronizer(runner)
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
				vertices[i].(*oneWayNode), vertices[j].(*oneWayNode),
				chans[counter], chans[counter+1])
			counter += 2
			log.Println("Channel", vertices[i].GetIndex(), "->", vertices[j].GetIndex(), "set up")
			log.Println("Channel", vertices[j].GetIndex(), "->", vertices[i].GetIndex(), "set up")
		}
	}
	for _, vertex := range vertices {
		vertex.(*oneWayNode).shuffleTopology()
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
					vertices[i].(*oneWayNode), vertices[j].(*oneWayNode),
					chans[0], chans[1])
				log.Println("Channel", vertices[i].GetIndex(), "->", vertices[j].GetIndex(), "set up")
				log.Println("Channel", vertices[j].GetIndex(), "->", vertices[i].GetIndex(), "set up")
			}
		}
	}
	for _, vertex := range vertices {
		vertex.(*oneWayNode).shuffleTopology()
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
			vertices[i].(*oneWayNode), vertices[j].(*oneWayNode),
			chans[counter], chans[counter+1])
		counter += 2
		log.Println("Channel", vertices[i].GetIndex(), "->", vertices[j].GetIndex(), "set up")
		log.Println("Channel", vertices[j].GetIndex(), "->", vertices[i].GetIndex(), "set up")
	}
	for _, vertex := range vertices {
		vertex.(*oneWayNode).shuffleTopology()
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
			if i < j {
				chans := getSynchronousChannels(2)
				addTwoWayConnection(
					vertices[i].(*oneWayNode), vertices[j].(*oneWayNode),
					chans[0], chans[1])
				log.Println("Channel", vertices[i].GetIndex(), "->", vertices[j].GetIndex(), "set up")
				log.Println("Channel", vertices[j].GetIndex(), "->", vertices[i].GetIndex(), "set up")
			}
		}
	}
	for _, vertex := range vertices {
		vertex.(*oneWayNode).shuffleTopology()
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
	for i := 0; i < a; i++ {
		for j := 0; j < b; j++ {
			if i-1 >= 0 {
				adjacencyList[i*b+j] = append(adjacencyList[i*b+j], (i-1)*b+j)
			}
			if j-1 >= 0 {
				adjacencyList[i*b+j] = append(adjacencyList[i*b+j], i*b+j-1)
			}
			if i+1 < a {
				adjacencyList[i*b+j] = append(adjacencyList[i*b+j], (i+1)*b+j)
			}
			if j+1 < b {
				adjacencyList[i*b+j] = append(adjacencyList[i*b+j], i*b+j+1)
			}
		}
	}
	return BuildGraphFromAdjacencyList(adjacencyList, GetRandomGenerator())
}

func BuildSynchronizedUndirectedMesh(a int, b int) ([]Node, Synchronizer) {
	vertices, runner := BuildUndirectedMesh(a, b)
	return vertices, asSynchronizer(runner)
}

func squareCheck(x int) bool {
	var int_root int = int(math.Sqrt(float64(x)))
	return (int_root * int_root) == x
}

func BuildSynchronizedTorus(n int) ([]Node, Synchronizer) {
	if n <= 0 {
		panic("Size cannot be <= 0.")
	}
	if !squareCheck(n) {
		panic("n is not a perfect square.")
	}

	if int(math.Sqrt(float64(n))) <= 4 {
		panic("Sqrt(n) should be greater than 4.")
	}

	log.Println("Building torus size", n)
	adjacencyList := make([][]int, n)

	for i := range adjacencyList {
		adjacencyList[i] = make([]int, 0)
	}

	side := int(math.Sqrt(float64(n)))

	for i := 0; i < side; i++ {
		for j := 0; j < side; j++ {
			adjacencyList[i*side+j] = append(adjacencyList[i*side+j], i*side+((j-1+side)%side))
			adjacencyList[i*side+j] = append(adjacencyList[i*side+j], i*side+((j+1+side)%side))
			adjacencyList[i*side+j] = append(adjacencyList[i*side+j], (((i-1+side)%side)*side + j))
			adjacencyList[i*side+j] = append(adjacencyList[i*side+j], (((i+1+side)%side)*side + j))
		}
	}

	return BuildSynchronizedGraphFromAdjacencyList(adjacencyList, GetRandomGenerator())
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
			addTwoWayConnection(vertices[i].(*oneWayNode), vertices[j].(*oneWayNode), chans[0], chans[1])
			log.Println("Channel", vertices[i].GetIndex(), "->", vertices[j].GetIndex(), "set up")
			log.Println("Channel", vertices[j].GetIndex(), "->", vertices[i].GetIndex(), "set up")
		}
	}
	if !oriented {
		for _, vertex := range vertices {
			vertex.(*oneWayNode).shuffleTopology()
		}
	}
	rand.Shuffle(len(vertices), func(i, j int) { vertices[i], vertices[j] = vertices[j], vertices[i] })
	return vertices, runner
}

func BuildSynchronizedHypercube(dim int, oriented bool) ([]Node, Synchronizer) {
	vertices, runner := BuildHypercube(dim, oriented)
	return vertices, asSynchronizer(runner)
}

func BuildCompleteGraphWithLoops(n int, oriented bool, indexGenerator Generator) ([]Node, Synchronizer) {
	vertices, synchronizer := BuildSynchronizedEmptyGraph(n, indexGenerator)
	chans := getSynchronousChannels(n * n)
	counter := 0
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			addOneWayConnection(vertices[i].(*oneWayNode), vertices[j].(*oneWayNode), chans[counter])
			counter += 1
			log.Println("Channel", vertices[i].GetIndex(), "->", vertices[j].GetIndex(), "set up")
		}
	}
	if !oriented {
		for _, vertex := range vertices {
			vertex.(*oneWayNode).shuffleTopology()
		}
	}
	return vertices, synchronizer
}

func BuildHamiltonianOrientedCompleteGraph(n int) ([]Node, Runner) {
	vertices, runner := BuildEmptyGraph(n, GetRandomGenerator())
	chans := getSynchronousChannels(n * (n - 1))
	counter := 0
	for i := 0; i < n; i++ {
		for j := 1; j < n; j++ {
			addOneWayConnection(vertices[i].(*oneWayNode), vertices[(i+j)%n].(*oneWayNode), chans[counter])
			counter += 1
			log.Println("Channel", vertices[i].GetIndex(), "->", vertices[(i+j)%n].GetIndex(), "set up")
		}
	}
	return vertices, runner
}

func BuildSynchronizedHamiltonianOrientedCompleteGraph(n int) ([]Node, Synchronizer) {
	vertices, runner := BuildHamiltonianOrientedCompleteGraph(n)
	return vertices, asSynchronizer(runner)
}
func BuildSynchronizedRandomConnectedGraphWithUniqueIndices(n int, m int) ([]Node, Synchronizer) {
	provider := func(i int) int {
		return i
	}
	weightedNodes, synchronizer := BuildSynchronizedRandomConnectedWeightedGraph(n, m, 1, GetUniquenessGenerator(provider))
	nodes := make([]Node, len(weightedNodes))
	for i, weightedNode := range weightedNodes {
		nodes[i] = weightedNode
	}
	return nodes, synchronizer
}
