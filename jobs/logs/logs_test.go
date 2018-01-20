package logs_test

import (
	"errors"
	"testing"

	"github.com/pcarranza/meeseeks-box/jobs/logs"
	stubs "github.com/pcarranza/meeseeks-box/testingstubs"
)

func Test_Logs(t *testing.T) {
	tt := []struct {
		name     string
		jobID    uint64
		logs     []string
		err      error
		expected logs.JobLog
	}{
		{
			name:  "with one line",
			jobID: 1,
			logs:  []string{"something"},
			err:   nil,
			expected: logs.JobLog{
				Output: "something",
			},
		},
		{
			name:  "with multiple lines",
			jobID: 2,
			logs:  []string{"something", "\n", "something else"},
			err:   nil,
			expected: logs.JobLog{
				Output: "something\nsomething else",
			},
		},
		{
			name:  "with errors",
			jobID: 3,
			logs:  []string{"bla"},
			err:   errors.New("something bad happened"),
			expected: logs.JobLog{
				Output: "bla",
				Error:  "something bad happened",
			},
		},
	}
	stubs.WithTmpDB(func() {
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
	stubs.WithTmpDB(func() {
		_, err := logs.Get(1)
		stubs.AssertEquals(t, logs.ErrNoLogsForJob, err)
	})
}

func Test_ErredOutJobHasError(t *testing.T) {
	stubs.WithTmpDB(func() {
		logs.SetError(1, errors.New("nasty error"))
		l, err := logs.Get(1)
		stubs.Must(t, "should be able to get a job with only an error", err)
		stubs.AssertEquals(t, logs.JobLog{Error: "nasty error"}, l)
	})
}
