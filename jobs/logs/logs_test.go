package logs_test

import (
	"errors"
	"testing"

	"github.com/gomeeseeks/meeseeks-box/jobs/logs"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	stubs "github.com/gomeeseeks/meeseeks-box/testingstubs"
)

func Test_Logs(t *testing.T) {
	tt := []struct {
		name     string
		jobID    uint64
		logs     []string
		err      error
		expected meeseeks.JobLog
	}{
		{
			name:  "with one line",
			jobID: 1,
			logs:  []string{"something"},
			err:   nil,
			expected: meeseeks.JobLog{
				Output: "something",
			},
		},
		{
			name:  "with multiple lines",
			jobID: 2,
			logs:  []string{"something", "something else"},
			err:   nil,
			expected: meeseeks.JobLog{
				Output: "something\nsomething else",
			},
		},
		{
			name:  "with errors",
			jobID: 3,
			logs:  []string{"bla"},
			err:   errors.New("something bad happened"),
			expected: meeseeks.JobLog{
				Output: "bla",
				Error:  "something bad happened",
			},
		},
	}
	stubs.WithTmpDB(func(_ string) {
		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				for _, line := range tc.logs {
					logs.Append(tc.jobID, line)
				}
				if tc.err != nil {
					logs.SetError(tc.jobID, tc.err)
				}
				actual, err := logs.Get(tc.jobID)
				stubs.Must(t, "could not get job logs back", err)
				stubs.AssertEquals(t, tc.expected, actual)
			})
		}
	})

}

func Test_GetLoglessJob(t *testing.T) {
	stubs.WithTmpDB(func(_ string) {
		_, err := logs.Get(1)
		stubs.AssertEquals(t, logs.ErrNoLogsForJob, err)
	})
}

func Test_ErredOutJobHasError(t *testing.T) {
	stubs.WithTmpDB(func(_ string) {
		logs.SetError(1, errors.New("nasty error"))
		l, err := logs.Get(1)
		stubs.Must(t, "should be able to get a job with only an error", err)
		stubs.AssertEquals(t, meeseeks.JobLog{Error: "nasty error"}, l)
	})
}
