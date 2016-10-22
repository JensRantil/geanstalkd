package inmemory

import (
	. "testing"

	"github.com/JensRantil/geanstalkd"
)

func TestJobHeapPriorityQueueImplementsJobPriorityQueue(t T) {
	var _ geanstalkd.JobPriorityQueue = JobHeapPriorityQueue{nil}
}
