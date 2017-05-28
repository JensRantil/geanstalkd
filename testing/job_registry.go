package testing

import (
	. "github.com/smartystreets/goconvey/convey"

	"github.com/JensRantil/geanstalkd"
)

// GenericJobRegistryTest tests that a JobRegistry behaves as a JobRegistry
// should.
func GenericJobRegistryTest(jr geanstalkd.JobRegistry) {
	testEmptyJobRegistry(jr)
	Convey("It should behave like a generic JobRegistry", func() {
		Convey("When adding a single job", func() {
			job := geanstalkd.Job{ID: testID}
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
				returnedJob, err := jr.GetByID(testID)
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
			_, err := jr.GetByID(testID)
			Convey("Then it should not exist", func() {
				So(err, ShouldEqual, geanstalkd.ErrJobMissing)
			})
		})
	})
}
