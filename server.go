package main

import (
	"container/heap"
	"errors"
	"sync"
	"time"

	"github.com/google/btree"
	//"golang.org/x/net/context"
)

var drainingError = errors.New("Draining.")

var errNotFound = errors.New("Not found")

const DEFAULT_BTREE_DEGREE = 16

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

func (s *server) Delete(id jobId) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// TODO: Avoid unnecessary allocation here. Probably using an interface.
	j := job{Id: id}

	item := s.jobById.Delete(jobIdJobBTreeItem(j))
	if item == nil {
		return errNotFound
	}
	heap.Remove(s.jobHeap, j.Index)

	return nil
}