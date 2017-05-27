package inmemory

import (
	"container/heap"

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

	if a, b := left.RunnableAt, right.RunnableAt; a != nil || b != nil {
		if a != nil && b != nil {
			return a.Before(*b)
		} else if a != nil {
			return true
		} else /*if b != nil*/ {
			return false
		}
	}

	if left.Priority < right.Priority {
		return true
	}

	if left.ID < right.ID {
		return true
	}

	return false
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
}

func NewJobHeapPriorityQueue() *JobHeapPriorityQueue {
	r := &JobHeapPriorityQueue{
		&jobHeapInterface{
			indexByJobID: make(map[geanstalkd.JobID]int),
		},
	}

	// Not sure this is needed for an empty heap. Doing it just in case.
	heap.Init(r.heap)

	return r
}

func (h *JobHeapPriorityQueue) Update(j *geanstalkd.Job) error {
	index, ok := h.heap.indexByJobID[j.ID]
	if !ok {
		return geanstalkd.ErrJobMissing
	}

	h.heap.jobs[index] = j
	heap.Fix(h.heap, index)

	return nil
}
func (h *JobHeapPriorityQueue) Pop() (*geanstalkd.Job, error) {
	if h.heap.Len() == 0 {
		return nil, geanstalkd.ErrEmptyQueue
	}
	job := heap.Pop(h.heap).(*geanstalkd.Job)
	delete(h.heap.indexByJobID, job.ID)
	return job, nil
}
func (h *JobHeapPriorityQueue) Peek() (*geanstalkd.Job, error) {
	if len(h.heap.jobs) == 0 {
		return nil, geanstalkd.ErrEmptyQueue
	}
	job := h.heap.jobs[0]
	return job, nil
}
func (h *JobHeapPriorityQueue) Push(j *geanstalkd.Job) error {
	if h.heap.HasID(j.ID) {
		return geanstalkd.ErrJobAlreadyExist
	}
	heap.Push(h.heap, j)
	return nil
}
func (h *JobHeapPriorityQueue) Remove(jid geanstalkd.JobID) error {
	index, ok := h.heap.indexByJobID[jid]
	if !ok {
		return geanstalkd.ErrJobMissing
	}

	heap.Remove(h.heap, index)
	delete(h.heap.indexByJobID, jid)

	return nil
}
