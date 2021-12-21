package lib

type Comparable interface {
	Less(Comparable) bool
}

type PriorityQueue []Comparable

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Empty() bool { return len(pq) == 0 }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Less(pq[j])
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(Comparable)
	*pq = append(*pq, item)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}
