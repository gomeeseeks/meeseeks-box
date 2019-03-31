package local_test

import (
	"errors"
	"testing"

	"gitlab.com/yakshaving.art/meeseeks-box/meeseeks"
	"gitlab.com/yakshaving.art/meeseeks-box/mocks"
	"gitlab.com/yakshaving.art/meeseeks-box/persistence"
)

func Test_Logs(t *testing.T) {
	tt := []struct {
		name     string
		jobID    uint64
		logs     []string
		getter   func(uint64) (meeseeks.JobLog, error)
		err      error
		expected meeseeks.JobLog
	}{
		{
			name:  "with one line",
			jobID: 1,
			logs:  []string{"something"},
			err:   nil,
			getter: func(jobID uint64) (meeseeks.JobLog, error) {
				return persistence.LogReader().Get(jobID)
			},
			expected: meeseeks.JobLog{
				Output: "something",
			},
		},
		{
			name:  "with multiple lines",
			jobID: 2,
			logs:  []string{"something", "something else"},
			err:   nil,
			getter: func(jobID uint64) (meeseeks.JobLog, error) {
				return persistence.LogReader().Get(jobID)
			},
			expected: meeseeks.JobLog{
				Output: "something\nsomething else",
			},
		},
		{
			name:  "with errors",
			jobID: 3,
			logs:  []string{"bla"},
			err:   errors.New("something bad happened"),
			getter: func(jobID uint64) (meeseeks.JobLog, error) {
				return persistence.LogReader().Get(jobID)
			},
			expected: meeseeks.JobLog{
				Output: "bla",
				Error:  "something bad happened",
			},
		},
		{
			name:  "with multiple lines but doing head",
			jobID: 4,
			logs:  []string{"something", "something else"},
			err:   nil,
			getter: func(jobID uint64) (meeseeks.JobLog, error) {
				return persistence.LogReader().Head(jobID, 1)
			},
			expected: meeseeks.JobLog{
				Output: "something",
			},
		},
		{
			name:  "with multiple lines but doing tail",
			jobID: 5,
			logs:  []string{"something", "something else"},
			err:   nil,
			getter: func(jobID uint64) (meeseeks.JobLog, error) {
				return persistence.LogReader().Tail(jobID, 1)
			},
			expected: meeseeks.JobLog{
				Output: "something else",
			},
		},
	}
	mocks.WithTmpDB(func(_ string) {
		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				lw := persistence.LogWriter()
				for _, line := range tc.logs {
					lw.Append(tc.jobID, line)
				}
				if tc.err != nil {
					lw.SetError(tc.jobID, tc.err)
				}
				actual, err := tc.getter(tc.jobID)
				mocks.Must(t, "could not get job logs back", err)
				mocks.AssertEquals(t, tc.expected, actual)
			})
		}
	})

}

func Test_GetLoglessJob(t *testing.T) {
	mocks.WithTmpDB(func(_ string) {
		_, err := persistence.LogReader().Get(1)

		mocks.AssertEquals(t, meeseeks.ErrNoLogsForJob, err)
	})
}

func Test_AppendEmptyStringToLogDoesntCreateALog(t *testing.T) {
	mocks.WithTmpDB(func(_ string) {
		lw := persistence.LogWriter()

		err := lw.Append(1, "")
		mocks.Must(t, "should be able to write an empty string to a log", err)

		err = lw.SetError(1, nil)
		mocks.Must(t, "should be able to get set a nil error in a log", err)

		_, err = persistence.LogReader().Get(1)
		mocks.AssertEquals(t, meeseeks.ErrNoLogsForJob, err)
	})
}

func Test_ErredOutJobHasError(t *testing.T) {
	mocks.WithTmpDB(func(_ string) {
		persistence.LogWriter().SetError(1, errors.New("nasty error"))

		l, err := persistence.LogReader().Get(1)

		mocks.Must(t, "should be able to get a job with only an error", err)
		mocks.AssertEquals(t, meeseeks.JobLog{Error: "nasty error"}, l)
	})
}
