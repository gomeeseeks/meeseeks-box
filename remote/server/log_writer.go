package server

import (
	"context"
	"errors"
	"io"

	"github.com/gomeeseeks/meeseeks-box/persistence"
	"github.com/gomeeseeks/meeseeks-box/remote/api"
	"github.com/sirupsen/logrus"
)

type logWriterServer struct{}

func (logWriterServer) Append(writer api.LogWriter_AppendServer) error {
	entry, err := writer.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		logrus.Errorf("got error receiving log entry: %s", err)
		return nil
	}

	return persistence.LogWriter().Append(entry.GetJobID(), entry.GetLine())
}

// SetError implements LogWriterServer SetError
func (logWriterServer) SetError(ctx context.Context, entry *api.ErrorLogEntry) (*api.Empty, error) {
	return &api.Empty{}, persistence.LogWriter().SetError(entry.GetJobID(), errors.New(entry.GetError()))
}
