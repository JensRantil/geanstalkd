package geanstalkd

import "errors"

// TODO: Split these up into separate variable groups.
var (
	ErrJobAlreadyExist = errors.New("Job already exists in registry.")
	ErrJobMissing      = errors.New("Job doesn't already exist.")
	ErrEmptyRegistry   = errors.New("Registry is empty.")
	ErrEmptyQueue      = errors.New("Queue is empty.")
)
