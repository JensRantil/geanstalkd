package main

import (
	"container/heap"
	"errors"
	"sync"
	"time"

	"github.com/google/btree"
	//"golang.org/x/net/context"
)

var errDraining = errors.New("Draining.")

var errNotFound = errors.New("Not found")

const DefaultBTreeDegree = 16

type server struct {
	jobByID *btree.BTree
	jobHeap *runnableAtJobHeap
	// TODO: Add stats

	// TODO: Investigate if a sync.RWMutex will be useful.
	lock sync.Mutex
	ids  <-chan (jobID)
}

const EXPECTED_NBR_OF_JOBS = 1

func newServer(ids <-chan (jobID)) *server {
	return &server{
		jobByID: btree.New(DefaultBTreeDegree),
		jobHeap: &runnableAtJobHeap{},
		ids:     ids,
	}
}

const UndefinedIndex = -1

func (s *server) BuildJob(pri priority, at time.Time, ttr time.Duration, jobdata []byte) job {
	return job{
		ID:         <-s.ids,
		RunnableAt: at,
		TimeToRun:  ttr,
		Body:       jobdata,
		Priority:   pri,
		Index:      UndefinedIndex,
	}
}

func (s *server) Add(j job) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.jobByID.Has(jobIDJobBTreeItem(j)) {
		// Sanity check.
		return errors.New("The job identifier already existed. This should never happen.")
	}

	s.jobByID.ReplaceOrInsert(jobIDJobBTreeItem(j))

	heap.Push(s.jobHeap, &j)

	return nil
}

func (s *server) Delete(id jobID) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// TODO: Avoid unnecessary allocation here. Probably using an interface.
	j := job{ID: id}

	item := s.jobByID.Delete(jobIDJobBTreeItem(j))
	if item == nil {
		return errNotFound
	}
	heap.Remove(s.jobHeap, j.Index)

	return nil
}
