package inmemory

import (
	"container/heap"

	"github.com/JensRantil/geanstalkd"
)

// Implementation of the container/heap.Interface. Supports zero value initialization.
type jobHeapInterface struct {
	jobs         []*geanstalkd.Job
	indexByJobId map[geanstalkd.JobID]int
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
		if a != nil && b != nil && a.Before(*b) {
			return true
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
	pq.indexByJobId[pq.jobs[i].ID] = i
	pq.indexByJobId[pq.jobs[j].ID] = j
}

func (pq *jobHeapInterface) Push(x interface{}) {
	n := len(pq.jobs)
	item := x.(*geanstalkd.Job)
	pq.indexByJobId[item.ID] = n
	pq.jobs = append(pq.jobs, item)
}

func (pq *jobHeapInterface) Pop() interface{} {
	old := pq.jobs
	n := len(old)
	item := old[n-1]
	delete(pq.indexByJobId, item.ID)

	// TODO: Support a shrinkable structure here.
	pq.jobs = old[0 : n-1]

	return item
}

// Bridge between geanstalkd.JobPriorityQueue and container/heap
// implementation. Supports zero value initialization.
type JobHeapPriorityQueue struct {
	heap *jobHeapInterface
}

func NewJobHeapPriorityQueue() *JobHeapPriorityQueue {
	return &JobHeapPriorityQueue{
		&jobHeapInterface{
			indexByJobId: make(map[geanstalkd.JobID]int),
		},
	}
}

func (h *JobHeapPriorityQueue) Update(j *geanstalkd.Job) error {
	index, ok := h.heap.indexByJobId[j.ID]
	if !ok {
		return geanstalkd.ErrJobMissing
	}

	h.heap.jobs[index] = j
	heap.Fix(h.heap, index)

	return nil
}
func (h *JobHeapPriorityQueue) Pop() (*geanstalkd.Job, error) {
	var err error
	item := heap.Pop(h.heap).(*geanstalkd.Job)
	if item == nil {
		err = geanstalkd.ErrEmptyQueue
	}
	return item, err
}
func (h *JobHeapPriorityQueue) Peek() (*geanstalkd.Job, error) {
	if len(h.heap.jobs) == 0 {
		return nil, geanstalkd.ErrEmptyQueue
	}
	job := h.heap.jobs[0]
	return job, nil
}
func (h *JobHeapPriorityQueue) Push(j *geanstalkd.Job) {
	heap.Push(h.heap, j)
}
func (h *JobHeapPriorityQueue) Remove(jid geanstalkd.JobID) error {
	index, ok := h.heap.indexByJobId[jid]
	if !ok {
		return geanstalkd.ErrJobMissing
	}

	heap.Remove(h.heap, index)
	delete(h.heap.indexByJobId, jid)

	return nil
}
