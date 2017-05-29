package inmemory

import (
	"container/heap"
	"sync"

	"github.com/JensRantil/geanstalkd"
)

type tubeHeapItem struct {
	item geanstalkd.JobPriorityQueue
	name geanstalkd.Tube
}

// Implementation of the container/heap.Interface. Supports zero value initialization.
type tubeHeapInterface struct {
	items       []*tubeHeapItem
	indexByTube map[geanstalkd.Tube]int
}

func (pq *tubeHeapInterface) HasID(id geanstalkd.Tube) bool {
	_, exists := pq.indexByTube[id]
	return exists
}

func (pq *tubeHeapInterface) Len() int { return len(pq.items) }

// Less compares two jobs in the heap slice first according to
// RunnableAt, then by Priority and last by Job ID. Notice that this
// covers both ready job heap as well as delayed job heap cases (but
// might make unnecessary comparisons, which is a future optimization
// to inject a job comparator if it speeds things up).
func (pq *tubeHeapInterface) Less(i, j int) bool {
	left, err := pq.items[i].item.Peek()
	if err != nil {
		if err == geanstalkd.ErrEmptyQueue {
			return false
		}
		// Unhandled error.
		panic(err)
	}
	right, err := pq.items[j].item.Peek()
	if err != nil {
		if err == geanstalkd.ErrEmptyQueue {
			return true
		}
		// Unhandled error.
		panic(err)
	}
	return geanstalkd.Less(*left, *right)
}

func (pq *tubeHeapInterface) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.indexByTube[pq.items[i].name] = i
	pq.indexByTube[pq.items[j].name] = j
}

func (pq *tubeHeapInterface) Push(x interface{}) {
	n := len(pq.items)
	item := x.(*tubeHeapItem)
	pq.indexByTube[item.name] = n
	pq.items = append(pq.items, item)
}

func (pq *tubeHeapInterface) Pop() interface{} {
	old := pq.items
	n := len(old)
	item := old[n-1]
	delete(pq.indexByTube, item.name)

	// TODO: Support a shrinkable structure here.
	pq.items = old[0 : n-1]

	return item
}

// TubeHeapPriorityQueue is an in-memory geanstalkd.TubePriorityQueue
// implementation backed by a heap. Use NewTubeHeapPriorityQueue to create one.
type TubeHeapPriorityQueue struct {
	heap *tubeHeapInterface
	lock sync.RWMutex
}

// NewTubeHeapPriorityQueue is returns a new TubeHeapPriorityQueue ready for
// immediate use.
func NewTubeHeapPriorityQueue() *TubeHeapPriorityQueue {
	r := &TubeHeapPriorityQueue{
		&tubeHeapInterface{
			indexByTube: make(map[geanstalkd.Tube]int),
		},
		sync.RWMutex{},
	}

	// Not sure this is needed for an empty heap. Doing it just in case.
	heap.Init(r.heap)

	return r
}

// FixByTube modifies a JobPriorityQueue previously modified.
func (h *TubeHeapPriorityQueue) FixByTube(tube geanstalkd.Tube) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	index, ok := h.heap.indexByTube[tube]
	if !ok {
		return geanstalkd.ErrQueueMissing
	}

	heap.Fix(h.heap, index)

	return nil
}

// Peek returns the job which would be returned if Pop() is called.
// geanstalkd.ErrEmptyQueue is returned if the queue is empty.
func (h *TubeHeapPriorityQueue) Peek() (geanstalkd.JobPriorityQueue, error) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	if len(h.heap.items) == 0 {
		return nil, geanstalkd.ErrEmptyQueue
	}
	return h.heap.items[0].item, nil
}

// Push adds a new job. If a job with the given ID already has been pushed,
// geanstalkd.ErrJobAlreadyExist is returned.
func (h *TubeHeapPriorityQueue) Push(name geanstalkd.Tube, queue geanstalkd.JobPriorityQueue) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	if h.heap.HasID(name) {
		return geanstalkd.ErrQueueAlreadyExist
	}
	heap.Push(h.heap, &tubeHeapItem{queue, name})
	return nil
}

// RemoveByTube removed a job with given ID previously pushed to this queue.
// geanstalkd.ErrJobMissing if a job with the given ID could not be found.
func (h *TubeHeapPriorityQueue) RemoveByTube(name geanstalkd.Tube) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	index, ok := h.heap.indexByTube[name]
	if !ok {
		return geanstalkd.ErrQueueMissing
	}

	heap.Remove(h.heap, index)
	delete(h.heap.indexByTube, name)

	return nil
}
