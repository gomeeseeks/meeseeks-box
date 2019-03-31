package jobs_test

import (
	"testing"

	"gitlab.com/yakshaving.art/meeseeks-box/meeseeks"
	"gitlab.com/yakshaving.art/meeseeks-box/mocks"
	"gitlab.com/yakshaving.art/meeseeks-box/persistence"
)

var req = meeseeks.Request{
	Command:  "mycommand",
	Username: "myself",
	Channel:  "general",
}

func Test_GettingAJobWorksWhenEmpty(t *testing.T) {
	mocks.Must(t, "failed to run tests", mocks.WithTmpDB(func(_ string) {
		_, err := persistence.Jobs().Get(1)
		mocks.AssertEquals(t, "no job could be found", err.Error())
	}))
}

func Test_GettingJobsWorksWhenEmpty(t *testing.T) {
	mocks.Must(t, "failed to run tests", mocks.WithTmpDB(func(_ string) {
		js, err := persistence.Jobs().Find(meeseeks.JobFilter{
			Limit: 10,
			Match: func(_ meeseeks.Job) bool {
				return true
			},
		})
		mocks.Must(t, "empty result of jobs find should not return an error", err)
		mocks.AssertEquals(t, 0, len(js))
	}))
}

func Test_CreatingAndThenGettingAJob(t *testing.T) {
	mocks.Must(t, "failed to run tests", mocks.WithTmpDB(func(_ string) {
		expected, err := persistence.Jobs().Create(req)
		mocks.Must(t, "Could not store a job: ", err)

		actual, err := persistence.Jobs().Get(expected.ID)
		mocks.Must(t, "Could not retrieve a job: ", err)

		mocks.AssertEquals(t, expected, actual)
	}))
}

func Test_MarkSuccessFul(t *testing.T) {
	mocks.Must(t, "failed to run tests", mocks.WithTmpDB(func(_ string) {
		job, err := persistence.Jobs().Create(req)
		mocks.Must(t, "Could not store a job: ", err)

		err = persistence.Jobs().Succeed(job.ID)
		mocks.Must(t, "could not set as successful", err)

		actual, err := persistence.Jobs().Get(job.ID)
		mocks.Must(t, "Could not retrieve a job: ", err)

		mocks.AssertEquals(t, actual.Status, meeseeks.JobSuccessStatus)
		if !actual.EndTime.After(job.StartTime) {
			t.Fatal("End time should be after start time")
		}
	}))
}
func Test_MarkFailedWorks(t *testing.T) {
	mocks.Must(t, "failed to run tests", mocks.WithTmpDB(func(_ string) {
		job, err := persistence.Jobs().Create(req)
		mocks.Must(t, "Could not store a job: ", err)

		running, err := persistence.Jobs().Get(job.ID)
		mocks.Must(t, "Could not retrieve a job: ", err)

		mocks.AssertEquals(t, running.Status, meeseeks.JobRunningStatus)

		err = persistence.Jobs().Fail(job.ID)
		mocks.Must(t, "could not set as successful", err)

		actual, err := persistence.Jobs().Get(job.ID)
		mocks.Must(t, "Could not retrieve a job: ", err)

		mocks.AssertEquals(t, actual.Status, meeseeks.JobFailedStatus)
		if !actual.EndTime.After(job.StartTime) {
			t.Fatal("End time should be after start time")
		}
	}))
}
func Test_FilterReturnsInOrder(t *testing.T) {
	mocks.Must(t, "failed to run tests", mocks.WithTmpDB(func(_ string) {
		persistence.Jobs().Create(req)
		persistence.Jobs().Create(req)
		persistence.Jobs().Create(req)

		latest, err := persistence.Jobs().Find(meeseeks.JobFilter{Limit: 2})
		if err != nil {
			t.Fatalf("Failed to get the latest jobs: %s", err)
		}
		mocks.AssertEquals(t, 2, len(latest))
		mocks.AssertEquals(t, uint64(3), latest[0].ID)
		mocks.AssertEquals(t, uint64(2), latest[1].ID)
	}))
}

func Test_FilterReturnsEnough(t *testing.T) {
	mocks.Must(t, "failed to run tests", mocks.WithTmpDB(func(_ string) {
		persistence.Jobs().Create(req)
		persistence.Jobs().Create(req)

		latest, err := persistence.Jobs().Find(meeseeks.JobFilter{Limit: 5})
		if err != nil {
			t.Fatalf("Failed to get the latest jobs: %s", err)
		}
		mocks.AssertEquals(t, 2, len(latest))
		mocks.AssertEquals(t, uint64(2), latest[0].ID)
		mocks.AssertEquals(t, uint64(1), latest[1].ID)
	}))
}

func TestFailRunningJobsLeavesNoJobRunning(t *testing.T) {
	mocks.Must(t, "failed to run tests", mocks.WithTmpDB(func(_ string) {
		persistence.Jobs().Create(req)
		persistence.Jobs().Create(req)
		persistence.Jobs().Create(req)

		mocks.Must(t, "Fail running jobs", persistence.Jobs().FailRunningJobs())

		running, err := persistence.Jobs().Find(meeseeks.JobFilter{
			Limit: 5,
			Match: func(j meeseeks.Job) bool {
				return j.Status == meeseeks.JobRunningStatus
			},
		})
		mocks.Must(t, "get running jobs", err)
		mocks.AssertEquals(t, 0, len(running))

		killed, err := persistence.Jobs().Find(meeseeks.JobFilter{
			Limit: 5,
			Match: func(j meeseeks.Job) bool {
				return j.Status == meeseeks.JobKilledStatus
			},
		})
		mocks.Must(t, "get killed jobs", err)
		mocks.AssertEquals(t, 3, len(killed))
	}))
}

func TestNullWorks(t *testing.T) {
	n := persistence.Jobs().Null(req)
	mocks.AssertEquals(t, uint64(0), n.ID)
	mocks.AssertEquals(t, meeseeks.JobRunningStatus, n.Status)
}
