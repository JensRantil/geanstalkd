package geanstalkd

import "errors"

// TODO: Split these up into separate variable groups.
var (
	ErrJobAlreadyExist   = errors.New("job already exists in registry")
	ErrJobMissing        = errors.New("job doesn't already exist")
	ErrQueueMissing      = errors.New("queue doesn't already exist")
	ErrEmptyRegistry     = errors.New("registry is empty")
	ErrEmptyQueue        = errors.New("queue is empty")
	ErrQueueAlreadyExist = errors.New("queue already exists")
)
