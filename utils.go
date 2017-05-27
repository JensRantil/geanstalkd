package geanstalkd

import "context"

// GenerateIds returns a channel with strictly monotonically increasing job
// IDs. Generation stops are soon ctx is done.
func GenerateIds(ctx context.Context) <-chan JobID {
	ids := make(chan JobID, 100)
	go func() {
		nextID := JobID(1)
		for {
			select {
			case ids <- nextID:
				nextID++
			case <-ctx.Done():
				return
			}
		}
	}()

	return ids
}
