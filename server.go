package main

import (
	"container/heap"
	"errors"
	"github.com/google/btree"
	"sync"
	"time"

	"golang.org/x/net/context"
)

var drainingError = errors.New("Draining.")

type counter uint
type jobId uint64
type priority uint64

var errNotFound = errors.New("Not found")

const DEFAULT_BTREE_DEGREE = 16

type job struct {
	Id         jobId
	RunnableAt time.Time
	TimeToRun  time.Duration
	Body       []byte
	Priority   priority

	// Index in the job heap.
	// TODO: Rename to better naming.
	Index int
}

// TODO: Support multiple tubes.
//type tube struct {
//	// TODO: Add stats
//}

type server struct {
	jobById *btree.BTree
	jobHeap *runnableAtJobHeap
	// TODO: Add stats

	// TODO: Investigate if a sync.RWMutex will be useful.
	lock sync.Mutex
	ids  <-chan (jobId)
}

const EXPECTED_NBR_OF_JOBS = 1

func newServer(ids <-chan (jobId)) *server {
	return &server{
		jobById: btree.New(DEFAULT_BTREE_DEGREE),
		jobHeap: &runnableAtJobHeap{},
		ids:     ids,
	}
}

const UNDEFINED_INDEX = -1

func (s *server) BuildJob(pri priority, at time.Time, ttr time.Duration, jobdata []byte) job {
	return job{
		Id:         <-s.ids,
		RunnableAt: at,
		TimeToRun:  ttr,
		Body:       jobdata,
		Priority:   pri,
		Index:      UNDEFINED_INDEX,
	}
}

func (s *server) Add(j job) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.jobById.Has(jobIdJobBTreeItem(j)) {
		// Sanity check.
		return errors.New("The job identifier already existed. This should never happen.")
	}

	s.jobById.ReplaceOrInsert(jobIdJobBTreeItem(j))

	heap.Push(s.jobHeap, &j)

	return nil
}
