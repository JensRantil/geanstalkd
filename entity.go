package geanstalkd

import (
	"time"
)

type JobID uint64
type Priority uint64

type Job struct {
	ID         JobID
	RunnableAt *time.Time
	TimeToRun  time.Duration
	Body       []byte
	Priority   Priority
}

type Tube struct {
	Name      string
	ReadyJobs TubePriorityQueue
}
