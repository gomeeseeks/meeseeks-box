package remote

import (
	"github.com/gomeeseeks/meeseeks-box/meeseeks/metrics"
)

// GRPCLogWriter contains all that's needed to create a remote log writer
type GRPCLogWriter struct {
	jobID uint64
}

// NewJobLogWriter returns a new GRPCLogWriter object
func NewJobLogWriter(jobID uint64) GRPCLogWriter {
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
