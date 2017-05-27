package geanstalkd

import (
	"errors"
	"sync"
	"time"
	//"context"
)

// Common `Server` errors.
var (
	ErrDraining = errors.New("server is draining. No new jobs can be added")
)

// Server is the facade through which all interactions to geanstalk go from the
// net layer.
type Server struct {
	Storage *LockService

	// TODO: Investigate if a sync.RWMutex will be useful.
	Ids <-chan (JobID)

	lock sync.Mutex
}

// BuildJob constructs a new job with an ID unique to this Server.
func (s *Server) BuildJob(pri Priority, at time.Time, ttr time.Duration, jobdata []byte) Job {
	return Job{
		ID:         <-s.Ids,
		RunnableAt: &at,
		TimeToRun:  ttr,
		Body:       jobdata,
		Priority:   pri,
	}
}

// Add adds a new job to this Server.
func (s *Server) Add(j *Job) error {
	// TODO: Delegate to `DelayService` if job is delayed.
	return s.Storage.Add(j)
}

// DeleteByID deletes a job with the given ID from this Server.
func (s *Server) DeleteByID(id JobID) error {
	return s.Storage.DeleteByID(id)
}
