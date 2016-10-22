package geanstalkd

import (
	"errors"
	"sync"
	"time"
	//"golang.org/x/net/context"
)

// Common `Server` errors.
var (
	ErrDraining = errors.New("Server is draining. No new jobs can be added.")
)

type Server struct {
	Storage *LockService

	// TODO: Investigate if a sync.RWMutex will be useful.
	Ids <-chan (JobID)

	lock sync.Mutex
}

func (s *Server) BuildJob(pri Priority, at time.Time, ttr time.Duration, jobdata []byte) Job {
	return Job{
		ID:         <-s.Ids,
		RunnableAt: &at,
		TimeToRun:  ttr,
		Body:       jobdata,
		Priority:   pri,
	}
}

func (s *Server) Add(j *Job) error {
	// TODO: Delegate to `DelayService` if job is delayed.
	return s.Storage.Add(j)
}

func (s *Server) Delete(id JobID) error {
	return s.Storage.Delete(id)
}
