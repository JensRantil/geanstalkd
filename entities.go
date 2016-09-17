package main

import (
	"time"
)

type consumerGroupId string

type consumerGroup struct {
	TubeHeap tubeHeap
}

type jobId uint64
type priority uint64

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
type tube struct {
	jobs  runnableAtJobHeap
	Index int
}
