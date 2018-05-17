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

func (remoteProvider) Reader() meeseeks.LogReader {
	return nullReader{}
}

type nullReader struct {
}

func (nullReader) Get(_ uint64) (meeseeks.JobLog, error) {
	return meeseeks.JobLog{}, nil
}

func (nullReader) Head(_ uint64, _ int) (meeseeks.JobLog, error) {
	return meeseeks.JobLog{}, nil
}

func (nullReader) Tail(_ uint64, _ int) (meeseeks.JobLog, error) {
	return meeseeks.JobLog{}, nil
}

// Writer implements LoggerProvider.Writer
func (remoteProvider) Writer() meeseeks.LogWriter {
	return remoteLogWriter{}
}

// remoteLogWriter contains all that's needed to create a remote log writer
type remoteLogWriter struct {
}

// Append adds a new line to the logs of the given Job
func (l remoteLogWriter) Append(jobID uint64, content string) error {
	metrics.LogLinesCount.Inc()
	return nil
}

// SetError sets the error message for the given Job
func (l remoteLogWriter) SetError(jobID uint64, jobErr error) error {
	return nil
}
