package geanstalkd

import (
	"errors"
	"time"
)

var (
	// ErrNoJobReady is returned when there is no job ready.
	ErrNoJobReady = errors.New("No job ready.")
	// ErrNoJobDelayed is returned when there is no delayed job ready.
	ErrNoJobDelayed = errors.New("No delayed job ready.")
)

// StorageService stores jobs. All operations are atomic in terms of storage.
// Calls to all of its functions are non-blocking.
type StorageService struct {
	Jobs       JobRegistry
	ReadyQueue JobPriorityQueue
	DelayQueue JobPriorityQueue
}

// Add adds a new job to the storage service. Returns ErrJobAlreadyExist if a
// job with the given ID has already been added.
func (s *StorageService) Add(j *Job) error {
	if err := s.Jobs.Insert(j); err != nil {
		return err
	}
	if j.RunnableAt != nil && time.Now().After(*j.RunnableAt) {
		s.ReadyQueue.Push(j)
	} else {
		s.DelayQueue.Push(j)
	}
	return nil
}

// Update updates a preexisting job's metadata. Returns ErrJobMissing if the
// job could not be found.
func (s *StorageService) Update(j *Job) error {
	if err := s.Jobs.Update(j); err != nil {
		return err
	}
	s.ReadyQueue.Update(j)
	s.DelayQueue.Update(j)
	return nil
}

// DeleteByID deletes a job with the given ID. Returns ErrJobMissing if the job
// could not be found.
func (s *StorageService) DeleteByID(id JobID) error {
	if err := s.Jobs.DeleteByID(id); err != nil {
		return err
	}
	s.ReadyQueue.RemoveByID(id)
	s.DelayQueue.RemoveByID(id)
	return nil
}

// Read queries a preexisting job by ID. Returns ErrJobMissing if the job could
// not be found.
func (s *StorageService) Read(id JobID) (*Job, error) {
	return s.Jobs.GetByID(id)
}

// PeekNextDelayed return the next ready job. Returns `ErrNoJobReady` if there
// are no jobs ready.
func (s *StorageService) PeekNextDelayed() (*Job, error) {
	item, err := s.DelayQueue.Peek()
	if err == ErrEmptyQueue {
		return nil, ErrNoJobDelayed
	}
	return item, err
}

// PopNextReady returns the next ready job. Returns ErrNoJobReady if no job is
// ready.
func (s *StorageService) PopNextReady() (*Job, error) {
	item, err := s.ReadyQueue.Pop()
	if err == ErrEmptyQueue {
		return item, ErrNoJobReady
	}
	return item, err
}

// TODO: Implement when adding tube support.
//func  (s *StorageService) Tubes() []Tube {
//}
