package geanstalkd

import (
	"time"
)

// JobID is the unique identifier of a job. Every new job gets assigned an
// incremented id.
type JobID uint64

// Priority if the priority for a job. When jobs are being polled for, jobs
// with a smaller priority value has higher priority than larger priority
// values.
type Priority uint64

// Job is the structure containing all the metadata for a job.
type Job struct {
	ID         JobID
	RunnableAt *time.Time
	TimeToRun  time.Duration
	Body       []byte
	Priority   Priority
}

// Copy creates a new copy of the job.
func (j Job) Copy() Job {
	return j
}
