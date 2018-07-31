package agent

import (
	"context"
	"time"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/remote/api"

	"github.com/sirupsen/logrus"
)

type grpcLogWriter struct {
	client         api.LogWriterClient
	timeoutSeconds time.Duration
}

// Append implements LogWritter.Append
func (g grpcLogWriter) Append(jobID uint64, content string) error {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeoutSeconds)
	defer cancel()

	w, e := g.client.Append(ctx)
	if e != nil {
		return e
	}

	logrus.Debugf("sending log job %d - '%s'", jobID, content)
	return w.Send(&api.LogEntry{
		JobID: jobID,
		Line:  content,
	})
}

// SetError implements LogWritter.SetError
func (g grpcLogWriter) SetError(jobID uint64, jobErr error) error {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeoutSeconds)
	defer cancel()

	logrus.Debugf("setting error for job %d - '%s'", jobID, jobErr)
	_, err := g.client.SetError(ctx, &api.ErrorLogEntry{
		JobID: jobID,
		Error: jobErr.Error(),
	})
	return err
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
