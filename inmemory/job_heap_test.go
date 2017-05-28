package inmemory

import (
	. "testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/JensRantil/geanstalkd/testing"
)

func TestJobHeapPriorityQueue(t *T) {
	t.Parallel()

	Convey("Given a fresh JobHeapPriorityQueue", t, func() {
		jpq := NewJobHeapPriorityQueue()
		testing.GenericJobPriorityQueueTest(jpq)
	})
}
