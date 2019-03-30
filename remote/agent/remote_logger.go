package agent

import (
	"context"
	"fmt"
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

	w, err := g.client.Append(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get a remote appender for job %d: %s", jobID, err)
	}

	logrus.Debugf("sending log job %d - '%s'", jobID, content)
	err = w.Send(&api.LogEntry{
		JobID: jobID,
		Line:  content,
	})
	if err != nil {
		logrus.Errorf("failed to send log to remote appender %d - '%s'", jobID, err)
	}
	return err
}

func (g grpcLogWriter) SetError(jobID uint64, err error) error {
	return nil
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
