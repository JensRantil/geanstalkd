package main

// TODO: Define the jobByRunnableTime
type runnableAtJobHeap []*job

func (pq runnableAtJobHeap) Len() int { return len(pq) }

func (pq runnableAtJobHeap) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	left := pq[i]
	right := pq[j]

	if left.Priority < right.Priority {
		return true
	}
	if left.Id < right.Id {
		return true
	}

	return false
}

func (pq runnableAtJobHeap) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *runnableAtJobHeap) Push(x interface{}) {
	n := len(*pq)
	item := x.(*job)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *runnableAtJobHeap) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.Index = -1 // for safety

	// TODO: Support a shrinkable structure here.
	*pq = old[0 : n-1]

	return item
}
