package inmemory

import (
	"container/heap"
	"sync"

	"github.com/JensRantil/geanstalkd"
)

// Implementation of the container/heap.Interface. Supports zero value initialization.
type jobHeapInterface struct {
	jobs         []*geanstalkd.Job
	indexByJobID map[geanstalkd.JobID]int
}

func (pq *jobHeapInterface) HasID(id geanstalkd.JobID) bool {
	_, exists := pq.indexByJobID[id]
	return exists
}

func (pq *jobHeapInterface) Len() int { return len(pq.jobs) }

// Less compares two jobs in the heap slice first according to
// RunnableAt, then by Priority and last by Job ID. Notice that this
// covers both ready job heap as well as delayed job heap cases (but
// might make unnecessary comparisons, which is a future optimization
// to inject a job comparator if it speeds things up).
func (pq *jobHeapInterface) Less(i, j int) bool {
	left := pq.jobs[i]
	right := pq.jobs[j]
	return geanstalkd.Less(*left, *right)
}

func (pq *jobHeapInterface) Swap(i, j int) {
	pq.jobs[i], pq.jobs[j] = pq.jobs[j], pq.jobs[i]
	pq.indexByJobID[pq.jobs[i].ID] = i
	pq.indexByJobID[pq.jobs[j].ID] = j
}

func (pq *jobHeapInterface) Push(x interface{}) {
	n := len(pq.jobs)
	item := x.(*geanstalkd.Job)
	pq.indexByJobID[item.ID] = n
	pq.jobs = append(pq.jobs, item)
}

func (pq *jobHeapInterface) Pop() interface{} {
	old := pq.jobs
	n := len(old)
	item := old[n-1]
	delete(pq.indexByJobID, item.ID)

	// TODO: Support a shrinkable structure here.
	pq.jobs = old[0 : n-1]

	return item
}

// JobHeapPriorityQueue is an in-memory geanstalkd.JobPriorityQueue
// implementation backed by a heap. Use NewJobHeapPriorityQueue to create one.
type JobHeapPriorityQueue struct {
	heap *jobHeapInterface
	lock sync.RWMutex
}

// NewJobHeapPriorityQueue returns a new JobHeapPriorityQueue ready for immediate use.
func NewJobHeapPriorityQueue() *JobHeapPriorityQueue {
	r := &JobHeapPriorityQueue{
		&jobHeapInterface{
			indexByJobID: make(map[geanstalkd.JobID]int),
		},
		sync.RWMutex{},
	}

	// Not sure this is needed for an empty heap. Doing it just in case.
	heap.Init(r.heap)

	return r
}

// Update modifies a job previously pushed.
func (h *JobHeapPriorityQueue) Update(j *geanstalkd.Job) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	index, ok := h.heap.indexByJobID[j.ID]
	if !ok {
		return geanstalkd.ErrJobMissing
	}

	h.heap.jobs[index] = j
	heap.Fix(h.heap, index)

	return nil
}

// Pop removes and returns the job with the highest priority.
// geanstalkd.ErrEmptyQueue is returned if the queue is empty.
func (h *JobHeapPriorityQueue) Pop() (*geanstalkd.Job, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if h.heap.Len() == 0 {
		return nil, geanstalkd.ErrEmptyQueue
	}
	job := heap.Pop(h.heap).(*geanstalkd.Job)
	delete(h.heap.indexByJobID, job.ID)
	return job, nil
}

// Peek returns the job which would be returned if Pop() is called.
// geanstalkd.ErrEmptyQueue is returned if the queue is empty.
func (h *JobHeapPriorityQueue) Peek() (*geanstalkd.Job, error) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	if len(h.heap.jobs) == 0 {
		return nil, geanstalkd.ErrEmptyQueue
	}
	job := h.heap.jobs[0]
	return job, nil
}

// Push adds a new job. If a job with the given ID already has been pushed,
// geanstalkd.ErrJobAlreadyExist is returned.
func (h *JobHeapPriorityQueue) Push(j *geanstalkd.Job) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	if h.heap.HasID(j.ID) {
		return geanstalkd.ErrJobAlreadyExist
	}
	heap.Push(h.heap, j)
	return nil
}

// RemoveByID removed a job with given ID previously pushed to this queue.
// geanstalkd.ErrJobMissing if a job with the given ID could not be found.
func (h *JobHeapPriorityQueue) RemoveByID(jid geanstalkd.JobID) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	index, ok := h.heap.indexByJobID[jid]
	if !ok {
		return geanstalkd.ErrJobMissing
	}

	heap.Remove(h.heap, index)
	delete(h.heap.indexByJobID, jid)

	return nil
}
