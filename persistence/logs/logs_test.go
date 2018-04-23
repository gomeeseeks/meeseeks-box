package logs_test

import (
	"errors"
	"testing"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/mocks"
	"github.com/gomeeseeks/meeseeks-box/persistence/logs"
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
	mocks.WithTmpDB(func(_ string) {
		for _, tc := range tt {
			logs.Configure(logs.LoggerConfig{
				LoggerType: "local",
			})
			t.Run(tc.name, func(t *testing.T) {
				lw := logs.GetJobLogWriter(tc.jobID)
				for _, line := range tc.logs {
					lw.Append(line)
				}
				if tc.err != nil {
					lw.SetError(tc.err)
				}
				actual, err := logs.GetJobLogReader(tc.jobID).Get()
				mocks.Must(t, "could not get job logs back", err)
				mocks.AssertEquals(t, tc.expected, actual)
			})
		}
	})

}

func Test_GetLoglessJob(t *testing.T) {
	mocks.WithTmpDB(func(_ string) {
		logs.Configure(logs.LoggerConfig{
			LoggerType: "local",
		})
		_, err := logs.GetJobLogReader(1).Get()
		mocks.AssertEquals(t, logs.ErrNoLogsForJob, err)
	})
}

func Test_ErredOutJobHasError(t *testing.T) {
	mocks.WithTmpDB(func(_ string) {
		logs.Configure(logs.LoggerConfig{
			LoggerType: "local",
		})
		logs.GetJobLogWriter(1).SetError(errors.New("nasty error"))
		l, err := logs.GetJobLogReader(1).Get()
		mocks.Must(t, "should be able to get a job with only an error", err)
		mocks.AssertEquals(t, meeseeks.JobLog{Error: "nasty error"}, l)
	})
}
