package main

import (
	"time"
)

type consumerGroupID string

type consumerGroup struct {
	TubeHeap tubeHeap
}

type jobID uint64
type priority uint64

type job struct {
	ID         jobID
	RunnableAt time.Time
	TimeToRun  time.Duration
	Body       []byte
	Priority   priority

	// Index in the job heap.
	// TODO: Rename to better naming.
	Index int
}

// TODO: Support multiple tubes.
type tube struct {
	jobs  runnableAtJobHeap
	Index int
}
