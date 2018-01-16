package jobs_test

import (
	"testing"

	"gitlab.com/mr-meeseeks/meeseeks-box/jobs"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/request"
	stub "gitlab.com/mr-meeseeks/meeseeks-box/testingstubs"
)

var req = request.Request{
	Command:  "mycommand",
	Username: "myself",
	Channel:  "general",
}

func Test_GettingAJobWorksWhenEmpty(t *testing.T) {
	stub.Must(t, "failed to run tests", stub.WithTmpDB(func() {
		_, err := jobs.Get(1)
		stub.AssertEquals(t, "no job could be found", err.Error())
	}))
}

func Test_GettingJobsWorksWhenEmpty(t *testing.T) {
	stub.Must(t, "failed to run tests", stub.WithTmpDB(func() {
		js, err := jobs.Find(jobs.JobFilter{
			Limit: 10,
			Match: func(_ jobs.Job) bool {
				return true
			},
		})
		stub.Must(t, "empty result of jobs find should not return an error", err)
		stub.AssertEquals(t, 0, len(js))
	}))
}

func Test_CreatingAndThenGettingAJob(t *testing.T) {
	stub.Must(t, "failed to run tests", stub.WithTmpDB(func() {
		expected, err := jobs.Create(req)
		stub.Must(t, "Could not store a job: ", err)

		actual, err := jobs.Get(expected.ID)
		stub.Must(t, "Could not retrieve a job: ", err)

		stub.AssertEquals(t, expected, actual)
	}))
}

func Test_MarkSuccessFul(t *testing.T) {
	stub.Must(t, "failed to run tests", stub.WithTmpDB(func() {
		job, err := jobs.Create(req)
		stub.Must(t, "Could not store a job: ", err)

		err = job.Finish(jobs.SuccessStatus)
		stub.Must(t, "could not set as successful", err)

		actual, err := jobs.Get(job.ID)
		stub.Must(t, "Could not retrieve a job: ", err)

		stub.AssertEquals(t, actual.Status, jobs.SuccessStatus)
		if !actual.EndTime.After(job.StartTime) {
			t.Fatal("End time should be after start time")
		}
	}))
}

func Test_MarkSuccessFulWithRunningEndStateFails(t *testing.T) {
	stub.Must(t, "failed to run tests", stub.WithTmpDB(func() {
		job, err := jobs.Create(req)
		stub.Must(t, "Could not store a job: ", err)

		err = job.Finish(jobs.RunningStatus)
		if err.Error() != "invalid status Running" {
			t.Fatalf("Wrong error %s", err)
		}
	}))
}

func Test_FilterReturnsInOrder(t *testing.T) {
	stub.Must(t, "failed to run tests", stub.WithTmpDB(func() {
		jobs.Create(req)
		jobs.Create(req)
		jobs.Create(req)

		latest, err := jobs.Find(jobs.JobFilter{Limit: 2})
		if err != nil {
			t.Fatalf("Failed to get the latest jobs: %s", err)
		}
		stub.AssertEquals(t, 2, len(latest))
		stub.AssertEquals(t, uint64(3), latest[0].ID)
		stub.AssertEquals(t, uint64(2), latest[1].ID)
	}))
}

func Test_FilterReturnsEnough(t *testing.T) {
	stub.Must(t, "failed to run tests", stub.WithTmpDB(func() {
		jobs.Create(req)
		jobs.Create(req)

		latest, err := jobs.Find(jobs.JobFilter{Limit: 5})
		if err != nil {
			t.Fatalf("Failed to get the latest jobs: %s", err)
		}
		stub.AssertEquals(t, 2, len(latest))
		stub.AssertEquals(t, uint64(2), latest[0].ID)
		stub.AssertEquals(t, uint64(1), latest[1].ID)
	}))
}
