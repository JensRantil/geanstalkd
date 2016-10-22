package geanstalkd

import (
	"errors"
	"time"
)

var (
	ErrNoJobReady   = errors.New("No job ready.")
	ErrNoJobDelayed = errors.New("No delayed job ready.")
)

// Stores jobs. All operations are atomic in terms of storage. No calls are blocking.
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

// Should fail if job doesn't exist.
func (s *StorageService) Update(j *Job) error {
	if err := s.Jobs.Update(j); err != nil {
		return err
	}
	s.ReadyQueue.Fix(j)
	s.DelayQueue.Fix(j)
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

// Return the next ready job. Returns `ErrNoJobReady` if there are no jobs ready.
func (s *StorageService) PeekNextDelayed(j *Job) (*Job, error) {
	item := s.DelayQueue.Peek()
	if item == nil {
		return nil, ErrNoJobDelayed
	}
	return item, nil
}

// TODO: Add `tube` as parameter.
func (s *StorageService) PopNextReady() (*Job, error) {
	item := s.DelayQueue.Pop()
	if item == nil {
		return nil, ErrNoJobDelayed
	}
	return item, nil
}

// TODO: Implement when adding tube support.
//func  (s *StorageService) Tubes() []Tube {
//}
