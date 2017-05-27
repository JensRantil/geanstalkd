package geanstalkd

import "io"

type Service interface {
	io.Closer
}

// TTRService keeps track of jobs that are reserved and puts them back in READY
// state if they are not `touch`ed or `delete`d.
type TTRService interface {
	Service

	Reserve(Job)
	Delete(Job)
	Touch(Job)
}

// DelayService converts delayed jobs to READY state when their delayed has
// passed.
//
// Two possible implementations of this:
//  - One in-memory. Probably using a heap or something.
//  - One that that continuously polls the lock service for next job that
//    should be transformed from DELAYED to READY. Will use much less memory.
//
// Uses `LockService` and `StatisticsService`.
type DelayService interface {
	Service
}

type StatisticsService interface {
}
