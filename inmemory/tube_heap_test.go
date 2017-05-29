package inmemory

import (
	. "testing"

	. "github.com/smartystreets/goconvey/convey"

	"fmt"
	"time"

	"github.com/JensRantil/geanstalkd"
)

func TestTubeHeapInterface(t *T) {
	var _ geanstalkd.TubePriorityQueue = (*TubeHeapPriorityQueue)(nil)
}

const testTubeName = "default"
const testID = geanstalkd.JobID(42)

func mustParse(format, s string) time.Time {
	time, err := time.Parse(format, s)
	if err != nil {
		panic(err)
	}
	return time
}

var (
	earlyTime = mustParse(time.Kitchen, "3:04PM")
	laterTime = mustParse(time.Kitchen, "4:05PM")
)

func TestTubeHeapPriorityQueue(t *T) {
	t.Parallel()

	Convey("Given a fresh TubeHeapPriorityQueue", t, func() {
		jpq := NewTubeHeapPriorityQueue()
		testEmptyHeapPriorityQueue(jpq)

		Convey("When adding a tube with a single job", func() {
			queue := NewJobHeapPriorityQueue()

			job := geanstalkd.Job{ID: testID}
			err := queue.Push(&job)
			So(err, ShouldBeNil)

			err = jpq.Push(testTubeName, queue)
			Convey("No error should be returned", func() {
				So(err, ShouldBeNil)
			})
			Convey("When removing the same queue", func() {
				err := jpq.RemoveByTube(testTubeName)
				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)
				})
				testEmptyHeapPriorityQueue(jpq)
			})
			Convey("When removing a missing queue", func() {
				missingTubeName := geanstalkd.Tube(testTubeName + "_some_random_prefix")
				err := jpq.RemoveByTube(missingTubeName)
				Convey("Then ErrQueueMissing should be returned", func() {
					So(err, ShouldEqual, geanstalkd.ErrQueueMissing)
				})
			})
			Convey("When peeking a queue", func() {
				peekedQueue, err := jpq.Peek()
				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)
				})
				Convey("Then the queue peeked should be the added queue", func() {
					So(peekedQueue, ShouldEqual, queue)
				})
			})
			Convey("When adding a second queue with lower priority", func() {
				lowPrioQueue := NewJobHeapPriorityQueue()
				err := jpq.Push("testTube", lowPrioQueue)
				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)
				})
				Convey("When peeking queue", func() {
					peekedQueue, err := jpq.Peek()
					Convey("Then no error should be returned", func() {
						So(err, ShouldBeNil)
					})
					Convey("Then it should return the high priority job", func() {
						So(peekedQueue, shouldHaveEqualTubeQueueFields, queue)
					})
				})
				Convey("When updating the low priority to be of higher priority (than first job)", func() {
					higherPriorityJob := geanstalkd.Job{ID: testID - 1}
					queue.Push(&higherPriorityJob)
					err := jpq.FixByTube("testTube")
					Convey("Then no error should be returned", func() {
						So(err, ShouldBeNil)
					})
					Convey("When peeking queue", func() {
						peekedQueue, err := jpq.Peek()
						Convey("Then no error should be returned", func() {
							So(err, ShouldBeNil)
						})
						Convey("Then it should return the new high priority job", func() {
							So(peekedQueue, shouldHaveEqualTubeQueueFields, queue)
						})
					})
				})
			})
			Convey("When adding a queue with the same tube name", func() {
				err = jpq.Push(testTubeName, queue)
				Convey("Then ErrJobAlreadyExist should be returned", func() {
					So(err, ShouldEqual, geanstalkd.ErrQueueAlreadyExist)
				})
			})
		})
	})
}

func shouldHaveEqualTubeQueueFields(actual interface{}, expected ...interface{}) string {
	a := actual.(geanstalkd.JobPriorityQueue)
	e := expected[0].(geanstalkd.JobPriorityQueue)
	ahead, _ := a.Peek()
	ehead, _ := e.Peek()
	if a != e {
		return fmt.Sprintf("%+v (head:%+v) != %+v (head:%+v)\n", a, ahead, e, ehead)
	}
	return ""
}

func testEmptyHeapPriorityQueue(jpq geanstalkd.TubePriorityQueue) {
	Convey("Then it should behave like an empty job priority queue", func() {
		Convey("When peeking the smallest tube", func() {
			_, err := jpq.Peek()
			Convey("Then ErrEmptyQueue should be returned", func() {
				So(err, ShouldEqual, geanstalkd.ErrEmptyQueue)
			})
		})
		Convey("When removing a tube", func() {
			err := jpq.RemoveByTube(testTubeName)
			Convey("Then ErrQueueMissing should be returned", func() {
				So(err, ShouldEqual, geanstalkd.ErrQueueMissing)
			})
		})
		Convey("When fixing a queue", func() {
			err := jpq.FixByTube(testTubeName)
			Convey("Then ErrQueueMissing should be returned", func() {
				So(err, ShouldEqual, geanstalkd.ErrQueueMissing)
			})
		})
	})
}
