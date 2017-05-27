package inmemory

import (
	"container/heap"
	. "testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/JensRantil/geanstalkd"
	"github.com/JensRantil/geanstalkd/testing"
)

func TestJobHeapPriorityQueue(t *T) {
	t.Parallel()

	Convey("Given a fresh JobHeapPriorityQueue", t, func() {
		jpq := NewJobHeapPriorityQueue()
		testing.TestJobPriorityQueue(jpq)
	})
}

func TestInternalHeap(t *T) {
	t.Parallel()

	h := &jobHeapInterface{
		jobs: []*geanstalkd.Job{
			{ID: 40},
			{ID: 39},
		},
		indexByJobID: make(map[geanstalkd.JobID]int),
	}
	heap.Init(h)

	if heap.Pop(h).(*geanstalkd.Job).ID != 39 {
		t.Fail()
	}
}
