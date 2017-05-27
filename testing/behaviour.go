package testing

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/kr/pretty"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/JensRantil/geanstalkd"
)

const TestId = 42

// TestJobRegistry tests that a JobRegistry behaves as a JobRegistry
// should.
func TestJobRegistry(jr geanstalkd.JobRegistry) {
	testEmptyJobRegistry(jr)
	Convey("It should behave like a generic JobRegistry", func() {
		Convey("When adding a single job", func() {
			job := geanstalkd.Job{ID: TestId}
			err := jr.Insert(&job)
			Convey("When adding the same job again", func() {
				err := jr.Insert(&job)
				Convey("Then ErrJobAlreadyExist should be returned", func() {
					So(err, ShouldEqual, geanstalkd.ErrJobAlreadyExist)
				})
			})
			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})
			Convey("When removing the same job", func() {
				err := jr.DeleteByID(job.ID)
				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)
				})
				testEmptyJobRegistry(jr)
			})
			Convey("When updating the job", func() {
				newPriority := job.Priority + 1
				err := jr.Update(&geanstalkd.Job{ID: job.ID, Priority: newPriority})
				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)
				})
				Convey("Then the job should be updated", func() {
					result, err := jr.GetByID(job.ID)
					So(result.Priority, ShouldEqual, newPriority)
					So(err, ShouldBeNil)
				})
			})
			Convey("Then the job's ID should be the largest", func() {
				largestID, err := jr.GetLargestID()
				So(err, ShouldBeNil)
				So(largestID, ShouldEqual, job.ID)
			})
			Convey("When querying for the job", func() {
				returnedJob, err := jr.GetByID(TestId)
				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)
				})
				Convey("Then it should be returned", func() {
					So(returnedJob.ID, ShouldEqual, job.ID)
				})
			})
			Convey("When adding a second job with higher ID", func() {
				job2 := geanstalkd.Job{ID: job.ID + 1}
				err := jr.Insert(&job2)
				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)
				})
				Convey("Then second job should be the highest", func() {
					largestID, err := jr.GetLargestID()
					So(err, ShouldBeNil)
					So(largestID, ShouldEqual, job2.ID)
				})
			})
		})
	})
}

func testEmptyJobRegistry(jr geanstalkd.JobRegistry) {
	Convey("It should behave like an empty registry", func() {
		Convey("When querying largest ID", func() {
			_, err := jr.GetLargestID()
			Convey("Then ErrEmptyRegistry should be returned", func() {
				So(err, ShouldEqual, geanstalkd.ErrEmptyRegistry)
			})
		})
		Convey("When updating a job", func() {
			err := jr.Update(&geanstalkd.Job{})
			Convey("Then ErrJobMissing is returned", func() {
				So(err, ShouldEqual, geanstalkd.ErrJobMissing)
			})
		})
		Convey("When deleting a job", func() {
			err := jr.DeleteByID(0)
			Convey("Then ErrJobMissing is returned", func() {
				So(err, ShouldEqual, geanstalkd.ErrJobMissing)
			})
		})
		Convey("When querying for a job", func() {
			_, err := jr.GetByID(TestId)
			Convey("Then it should not exist", func() {
				So(err, ShouldEqual, geanstalkd.ErrJobMissing)
			})
		})
	})
}

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

// TestJobPriorityQueue tests that a JobRegistry behaves as a
// JobRegistry should.
func TestJobPriorityQueue(jpq geanstalkd.JobPriorityQueue) {
	Convey("It should behave like a generic JobPriorityQueue", func() {
		testEmptyJobPriorityQueue(jpq)

		Convey("When adding a job", func() {
			job := geanstalkd.Job{ID: 55}
			err := jpq.Push(&job)
			Convey("No error should be returned", func() {
				So(err, ShouldBeNil)
			})
			Convey("When popping a job", func() {
				poppedJob, err := jpq.Pop()
				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)
				})
				Convey("Then the job popped should be the added job", func() {
					So(*poppedJob, shouldHaveEqualJobFields, job)
				})
				testEmptyJobPriorityQueue(jpq)
			})
			Convey("When removing the job", func() {
				err := jpq.Remove(job.ID)
				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)
				})
				testEmptyJobPriorityQueue(jpq)
			})
			Convey("When removing a job missing", func() {
				missingId := job.ID + 1
				err := jpq.Remove(missingId)
				Convey("Then ErrJobMissing should be returned", func() {
					So(err, ShouldEqual, geanstalkd.ErrJobMissing)
				})
			})
			Convey("When peeking a job", func() {
				peekedJob, err := jpq.Peek()
				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)
				})
				Convey("Then the job popped should be the added job", func() {
					So(*peekedJob, shouldHaveEqualJobFields, job)
				})
			})
			Convey("When adding a second job with lower priority", func() {
				lowerPrioJob := geanstalkd.Job{ID: job.ID + 1}
				err := jpq.Push(&lowerPrioJob)
				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)
				})
				Convey("When peeking queue", func() {
					peekedJob, err := jpq.Peek()
					Convey("Then no error should be returned", func() {
						So(err, ShouldBeNil)
					})
					Convey("Then it should return the high priority job", func() {
						So(*peekedJob, shouldHaveEqualJobFields, job)
					})
				})
				Convey("When popping a job", func() {
					peekedJob, err := jpq.Pop()
					Convey("Then no error should be returned", func() {
						So(err, ShouldBeNil)
					})
					Convey("Then it should return the high priority job", func() {
						So(*peekedJob, shouldHaveEqualJobFields, job)
					})
				})
				Convey("When updating the low priority to be of higher priority (than first job)", func() {
					modifiedLowerPrioJob := lowerPrioJob.Copy()
					now := time.Now()
					modifiedLowerPrioJob.RunnableAt = &now
					err := jpq.Update(&modifiedLowerPrioJob)
					Convey("Then no error should be returned", func() {
						So(err, ShouldBeNil)
					})
					Convey("When peeking queue", func() {
						peekedJob, err := jpq.Peek()
						Convey("Then no error should be returned", func() {
							So(err, ShouldBeNil)
						})
						Convey("Then it should return the new high priority job", func() {
							So(*peekedJob, shouldHaveEqualJobFields, modifiedLowerPrioJob)
						})
					})
					Convey("When popping a job", func() {
						poppedJob, err := jpq.Peek()
						Convey("Then no error should be returned", func() {
							So(err, ShouldBeNil)
						})
						Convey("Then it should return the new high priority job", func() {
							So(*poppedJob, shouldHaveEqualJobFields, modifiedLowerPrioJob)
						})
					})
				})
			})
			Convey("When adding a job with the same ID", func() {
				err := jpq.Push(&job)
				Convey("Then ErrJobAlreadyExist should be returned", func() {
					So(err, ShouldEqual, geanstalkd.ErrJobAlreadyExist)
				})
			})
		})

		orderedJobs := []orderedTestJob{
			{geanstalkd.Job{ID: TestId, RunnableAt: &earlyTime}, "job with early runnable time"},
			{geanstalkd.Job{ID: TestId - 1, RunnableAt: &laterTime}, "job with later runnable time"},
			{geanstalkd.Job{ID: TestId - 3}, "regular job"},
			{geanstalkd.Job{ID: TestId - 2, Priority: 1}, "job with lower priority"},
			{geanstalkd.Job{ID: TestId, Priority: 1}, "job with higher ID"},
		}

		for i, job := range orderedJobs {
			earlyJob := job
			if i+1 == len(orderedJobs) {
				continue
			}
			laterJob := orderedJobs[i+1]

			// Test pushing the items in two different orders.
			testOrdered(jpq, earlyJob, laterJob, true)
			testOrdered(jpq, laterJob, earlyJob, false)
		}
	})
}

type orderedTestJob struct {
	job  geanstalkd.Job
	desc string
}

func testOrdered(jpq geanstalkd.JobPriorityQueue, j1, j2 orderedTestJob, firstHighPrio bool) {
	var highestPrio orderedTestJob
	var firstPrio, secondPrio string
	if firstHighPrio {
		highestPrio = j1
		firstPrio = "high prio"
		secondPrio = "low prio"
	} else {
		highestPrio = j2
		firstPrio = "low prio"
		secondPrio = "high prio"
	}

	Convey(fmt.Sprintf("When adding (the %s job) a %s", firstPrio, j1.desc), func() {
		err := jpq.Push(&j1.job)
		Convey("Then no error should be returned", func() {
			So(err, ShouldBeNil)
		})
		Convey(fmt.Sprintf("When adding (the %s job) a %s", secondPrio, j2.desc), func() {
			err := jpq.Push(&j2.job)
			Convey("When peeking", func() {
				job, err := jpq.Peek()
				testJob(err, *job, highestPrio)
			})
			Convey("When popping", func() {
				job, err := jpq.Pop()
				testJob(err, *job, highestPrio)
			})
			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func testJob(err error, job geanstalkd.Job, expected orderedTestJob) {
	Convey("Then no error should be returned", func() {
		So(err, ShouldBeNil)
	})
	Convey(fmt.Sprintf("Then the returned job should be the %s", expected.desc), func() {
		So(job, shouldHaveEqualJobFields, expected.job)
	})
}

func shouldHaveEqualJobFields(actual interface{}, expected ...interface{}) string {
	a := actual.(geanstalkd.Job)
	e := expected[0].(geanstalkd.Job)
	if !reflect.DeepEqual(a, e) {
		return fmt.Sprintf("Fields differing between actual and expected:\n%s", strings.Join(pretty.Diff(a, e), "\n"))
	}
	return ""
}

func testEmptyJobPriorityQueue(jpq geanstalkd.JobPriorityQueue) {
	Convey("Then it should behave like an empty job priority queue", func() {
		Convey("When popping an item", func() {
			_, err := jpq.Pop()
			Convey("Then ErrEmptyQueue should be returned", func() {
				So(err, ShouldEqual, geanstalkd.ErrEmptyQueue)
			})
		})
		Convey("When peeking an item", func() {
			_, err := jpq.Peek()
			Convey("Then ErrEmptyQueue should be returned", func() {
				So(err, ShouldEqual, geanstalkd.ErrEmptyQueue)
			})
		})
		Convey("When removing a job", func() {
			err := jpq.Remove(TestId)
			Convey("Then ErrJobMissing should be returned", func() {
				So(err, ShouldEqual, geanstalkd.ErrJobMissing)
			})
		})
		Convey("When updating a job", func() {
			err := jpq.Update(&geanstalkd.Job{ID: TestId})
			Convey("Then ErrJobMissing should be returned", func() {
				So(err, ShouldEqual, geanstalkd.ErrJobMissing)
			})
		})
	})
}
