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
	for {
		entry, err := writer.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logrus.Errorf("got error receiving log entry: %s", err)
			break
		}
		if err := persistence.LogWriter().Append(entry.GetJobID(), entry.GetLine()); err != nil {
			logrus.Errorf("got error receiving log entry: %s", err)
			break
		}
	}
	return writer.SendAndClose(&api.Empty{})
}

// SetError implements LogWriterServer SetError
func (logWriterServer) SetError(ctx context.Context, entry *api.ErrorLogEntry) (*api.Empty, error) {
	// Do I even need this?
	return &api.Empty{}, persistence.LogWriter().SetError(entry.GetJobID(), errors.New(entry.GetError()))
}
