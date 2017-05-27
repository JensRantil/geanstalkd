package geanstalkd

// A JobRegistry stores and queries jobs.
type JobRegistry interface {
	// Insert stores a new job in the registry. If a job with the same job
	// identifier already exist, `ErrJobAlreadyExist` returned.
	Insert(*Job) error

	// Update the registry to reflect possible changes made to Job. Returns
	// `ErrJobMissing` if the Job could not be found.
	Update(*Job) error

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

// A JobPriorityQueue is a queue which orders jobs according to a specific
// priority. The queue MAY be backed by a heap, but could equally be backed by
// a B-tree/LSM on disk.
//
// Jobs have the following priority:
//
// 1. Jobs with RunnableAt. If two jobs have it defined, the one with the
//    earliest value has higher precedence.
// 2. Jobs with the lower priority value are returned before higher priority
//    value.
// 3. Jobs with lower ID have higher precedence.
type JobPriorityQueue interface {
	// Notify the priority queue that the Job's priority might have changed and
	// that internal datastructures must be updated to reflect that. Returns
	// `ErrJobMissing` if the job was not in the queue.
	Update(*Job) error

	// Remove and return the Job with the highest priority. Returns
	// `ErrEmptyQueue` if there are no jobs in the queue.
	Pop() (*Job, error)

	// Return the highest priority Job. Return nil if there are no jobs in the
	// queue. Returns `ErrEmptyQueue` if there are no jobs in the queue.
	Peek() (*Job, error)

	// Put a new Job on the queue. Returns `ErrJobAlreadyExist` if job already
	// exists.
	Push(*Job) error

	// Remove a Job from the queue with a specific ID.  Returns `ErrJobMissing`
	// if the job was not in the queue.
	//
	// TODO: Rename to `RemoveByID` to make this using same nomenclature as
	// other functions.
	Remove(JobID) error
}
