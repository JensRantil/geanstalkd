package geanstalkd

type JobRegistry interface {
	// Insert stores a new job in the registry. If a job with the same job
	// identifier already exist, `ErrJobAlreadyExist` returned.
	Insert(Job) error

	Update(Job) error

	// GetByID returns the job with a given job identifier. If the job can't be
	// found, it returns `ErrJobMissing`.
	GetByID(JobID) (*Job, error)

	// DeleteByID deletes a job by job identitifer. If the job can't be found,
	// it returns `ErrJobMissing`.
	DeleteByID(JobID) error

	// GetLargestID returns the JobID of the job with the highest ID. This
	// method is only called on initialization. Returns `ErrEmptyRegistry` if
	// the registry contains no jobs.
	GetLargestID() (JobID, error)
}

type JobPriorityQueue interface {
	Update(Job)
	Pop() *Job
	Peek() *Job
	Push(Job)
	Remove(JobID)
}
