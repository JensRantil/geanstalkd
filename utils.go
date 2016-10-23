package geanstalkd

import "context"

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
