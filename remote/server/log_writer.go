package server

import (
	"context"
	"errors"
	"io"

	"gitlab.com/yakshaving.art/meeseeks-box/persistence"
	"gitlab.com/yakshaving.art/meeseeks-box/remote/api"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type logWriterServer struct{}

func (logWriterServer) Append(writer api.LogWriter_AppendServer) error {
Loop:
	for {
		entry, err := writer.Recv()
		if err == io.EOF {
			logrus.Errorf("got EOF receiving log entry")
			break Loop
		}

		errCode := status.Code(err)
		switch errCode {
		case codes.OK:
			logrus.Debugf("logline received is OK, continuing")

		case codes.Canceled, codes.DeadlineExceeded:
			logrus.Infof("timed out waiting for new log lines, breaking out")
			break Loop

		default:
			logrus.Errorf("logger erred out with: %v - %s", errCode, err)
			break Loop

		}

		err = persistence.LogWriter().Append(entry.GetJobID(), entry.GetLine())
		if err != nil {
			logrus.Errorf("got error receiving log entry: %s", err)
		} else {
			logrus.Debugf("appended new log line to job %d", entry.GetJobID())
		}

		break Loop
	}
	return writer.SendAndClose(&api.Empty{})
}

// SetError implements LogWriterServer SetError
func (logWriterServer) SetError(ctx context.Context, entry *api.ErrorLogEntry) (*api.Empty, error) {
	// Do I even need this?
	return &api.Empty{}, persistence.LogWriter().SetError(entry.GetJobID(), errors.New(entry.GetError()))
}
