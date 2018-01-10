package jobs_test

import (
	"testing"

	"strings"

	"gitlab.com/mr-meeseeks/meeseeks-box/jobs"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/request"
	stub "gitlab.com/mr-meeseeks/meeseeks-box/testingstubs"
)

var req = request.Request{
	Command:  "mycommand",
	Username: "myself",
	Channel:  "general",
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

		err = jobs.Finish(job.ID, jobs.SuccessStatus)
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

		err = jobs.Finish(job.ID, jobs.RunningStatus)
		if !strings.HasPrefix(err.Error(), "invalid end state") {
			t.Fatalf("Wrong error %s", err)
		}
	}))
}

func Test_LatestReturnsInOrder(t *testing.T) {
	stub.Must(t, "failed to run tests", stub.WithTmpDB(func() {
		jobs.Create(req)
		jobs.Create(req)
		jobs.Create(req)

		latest, err := jobs.Latest(2)
		if err != nil {
			t.Fatalf("Failed to get the latest jobs: %s", err)
		}
		stub.AssertEquals(t, 2, len(latest))
		stub.AssertEquals(t, uint64(3), latest[0].ID)
		stub.AssertEquals(t, uint64(2), latest[1].ID)
	}))
}

func Test_LatestReturnsEnough(t *testing.T) {
	stub.Must(t, "failed to run tests", stub.WithTmpDB(func() {
		jobs.Create(req)
		jobs.Create(req)

		latest, err := jobs.Latest(5)
		if err != nil {
			t.Fatalf("Failed to get the latest jobs: %s", err)
		}
		stub.AssertEquals(t, 2, len(latest))
		stub.AssertEquals(t, uint64(2), latest[0].ID)
		stub.AssertEquals(t, uint64(1), latest[1].ID)
	}))
}
