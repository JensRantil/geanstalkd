package geanstalkd

import (
	"errors"
	"time"
)

var (
	ErrNoJobReady   = errors.New("No job ready.")
	ErrNoJobDelayed = errors.New("No delayed job ready.")
)

// StorageService stores jobs. All operations are atomic in terms of storage.
// Calls to all of its functions are non-blocking.
type StorageService struct {
	Jobs       JobRegistry
	ReadyQueue JobPriorityQueue
	DelayQueue JobPriorityQueue
}

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

func (s *StorageService) Delete(id JobID) error {
	if err := s.Jobs.DeleteByID(id); err != nil {
		return err
	}
	s.ReadyQueue.Remove(id)
	s.DelayQueue.Remove(id)
	return nil
}

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
