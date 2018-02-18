package jobs_test

import (
	"testing"

	"github.com/gomeeseeks/meeseeks-box/jobs"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/mocks"
)

var req = meeseeks.Request{
	Command:  "mycommand",
	Username: "myself",
	Channel:  "general",
}

func Test_GettingAJobWorksWhenEmpty(t *testing.T) {
	mocks.Must(t, "failed to run tests", mocks.WithTmpDB(func(_ string) {
		_, err := jobs.Get(1)
		mocks.AssertEquals(t, "no job could be found", err.Error())
	}))
}

func Test_GettingJobsWorksWhenEmpty(t *testing.T) {
	mocks.Must(t, "failed to run tests", mocks.WithTmpDB(func(_ string) {
		js, err := jobs.Find(jobs.JobFilter{
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
		expected, err := jobs.Create(req)
		mocks.Must(t, "Could not store a job: ", err)

		actual, err := jobs.Get(expected.ID)
		mocks.Must(t, "Could not retrieve a job: ", err)

		mocks.AssertEquals(t, expected, actual)
	}))
}

func Test_MarkSuccessFul(t *testing.T) {
	mocks.Must(t, "failed to run tests", mocks.WithTmpDB(func(_ string) {
		job, err := jobs.Create(req)
		mocks.Must(t, "Could not store a job: ", err)

		err = jobs.Finish(job.ID, jobs.SuccessStatus)
		mocks.Must(t, "could not set as successful", err)

		actual, err := jobs.Get(job.ID)
		mocks.Must(t, "Could not retrieve a job: ", err)

		mocks.AssertEquals(t, actual.Status, jobs.SuccessStatus)
		if !actual.EndTime.After(job.StartTime) {
			t.Fatal("End time should be after start time")
		}
	}))
}

func Test_MarkSuccessFulWithRunningEndStateFails(t *testing.T) {
	mocks.Must(t, "failed to run tests", mocks.WithTmpDB(func(_ string) {
		job, err := jobs.Create(req)
		mocks.Must(t, "Could not store a job: ", err)

		err = jobs.Finish(job.ID, jobs.RunningStatus)
		if err.Error() != "invalid status Running" {
			t.Fatalf("Wrong error %s", err)
		}
	}))
}

func Test_FilterReturnsInOrder(t *testing.T) {
	mocks.Must(t, "failed to run tests", mocks.WithTmpDB(func(_ string) {
		jobs.Create(req)
		jobs.Create(req)
		jobs.Create(req)

		latest, err := jobs.Find(jobs.JobFilter{Limit: 2})
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
		jobs.Create(req)
		jobs.Create(req)

		latest, err := jobs.Find(jobs.JobFilter{Limit: 5})
		if err != nil {
			t.Fatalf("Failed to get the latest jobs: %s", err)
		}
		mocks.AssertEquals(t, 2, len(latest))
		mocks.AssertEquals(t, uint64(2), latest[0].ID)
		mocks.AssertEquals(t, uint64(1), latest[1].ID)
	}))
}
