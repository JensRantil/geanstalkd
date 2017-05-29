package geanstalkd

// Less implements the Job comparison of which Job to run first. A job which is
// "Less" is processed before another job.
//
// Jobs are first compared according to RunnableAt, then by Priority and last
// by Job ID. Notice that this covers both ready job heap as well as delayed
// job heap cases (but might make unnecessary comparisons, which is a future
// optimization to inject a job comparator if it speeds things up).
func Less(left, right Job) bool {
	if a, b := left.RunnableAt, right.RunnableAt; a != nil || b != nil {
		if a != nil && b != nil {
			return a.Before(*b)
		} else if a != nil {
			return true
		} else /*if b != nil*/ {
			return false
		}
	}

	if left.Priority < right.Priority {
		return true
	}

	if left.ID < right.ID {
		return true
	}

	return false
}
