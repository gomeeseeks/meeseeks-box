package remote

import (
	"errors"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/metrics"
)

// ErrNoLogsForJob is returned when we try to extract the logs of a non existing job
var ErrNoLogsForJob = errors.New("No logs for job")

// LogWriter is an interface to write job logs
type LogWriter interface {
	Append(content string) error
	SetError(jobErr error) error
}

// GRPCLogWriter contains all that's needed to create a remote log writer
type GRPCLogWriter struct {
	jobID uint64
}

// NewJobLogWriter returns a new GRPCLogWriter object
func NewJobLogWriter(jobID uint64) LogWriter {
	return GRPCLogWriter{
		jobID: jobID,
	}
}

// Append adds a new line to the logs of the given Job
func (l GRPCLogWriter) Append(content string) error {
	metrics.LogLinesCount.Inc()
	return nil
}

// SetError sets the error message for the given Job
func (l GRPCLogWriter) SetError(jobErr error) error {
	return nil
}
