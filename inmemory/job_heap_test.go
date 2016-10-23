package inmemory

import (
	. "testing"

	"github.com/JensRantil/geanstalkd"
)

func TestJobHeapPriorityQueueImplementsJobPriorityQueue(t *T) {
	t.Parallel()
	var _ geanstalkd.JobPriorityQueue = (*JobHeapPriorityQueue)(nil)
}
