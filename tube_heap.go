package main

// TODO: Define the jobByRunnableTime
type tubeHeap []*tube

func (pq tubeHeap) Len() int { return len(pq) }

func (pq tubeHeap) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	left := pq[i]
	right := pq[j]

	if len(right.jobs) == 0 {
		return true
	}

	if len(left.jobs) == 0 {
		return false
	}

	leftFirstJob := left.jobs[0]
	rightFirstJob := right.jobs[0]

	if leftFirstJob.Priority < rightFirstJob.Priority {
		return true
	}
	if leftFirstJob.ID < rightFirstJob.ID {
		return true
	}

	return false
}

func (pq tubeHeap) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *tubeHeap) Push(x interface{}) {
	n := len(*pq)
	item := x.(*tube)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *tubeHeap) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.Index = -1 // for safety

	// TODO: Support a shrinkable structure here.
	*pq = old[0 : n-1]

	return item
}
