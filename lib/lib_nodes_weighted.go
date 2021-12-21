package lib

type WeightedGraphNode interface {
	Node
	GetInWeights() []int
	GetOutWeights() []int
}
