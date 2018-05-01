package remote

import (
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/metrics"
)

type remoteProvider struct{}

// New returns a new Remote Logger Provider
func New() meeseeks.LoggerProvider {
	return remoteProvider{}
}

func (remoteProvider) Reader(_ uint64) meeseeks.LogReader {
	return nullReader{}
}

type nullReader struct {
}

func (nullReader) Get() (meeseeks.JobLog, error) {
	return meeseeks.JobLog{}, nil
}

func (nullReader) Head(_ int) (meeseeks.JobLog, error) {
	return meeseeks.JobLog{}, nil
}

func (nullReader) Tail(_ int) (meeseeks.JobLog, error) {
	return meeseeks.JobLog{}, nil
}

// Writer implements LoggerProvider.Writer
func (remoteProvider) Writer(j uint64) meeseeks.LogWriter {
	return remoteLogWriter{
		jobID: j,
	}
}

// remoteLogWriter contains all that's needed to create a remote log writer
type remoteLogWriter struct {
	jobID uint64
}

// Append adds a new line to the logs of the given Job
func (l remoteLogWriter) Append(content string) error {
	metrics.LogLinesCount.Inc()
	return nil
}

// SetError sets the error message for the given Job
func (l remoteLogWriter) SetError(jobErr error) error {
	return nil
}
