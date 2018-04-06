package server

import (
	"context"
	"fmt"
	"github.com/gomeeseeks/meeseeks-box/jobs/logs"
	pb "github.com/gomeeseeks/meeseeks-box/remote/api"
	"github.com/sirupsen/logrus"
	"io"
	"time"
)

// CommandLoggerServer implements the remote logger interface
type CommandLoggerServer struct{}

// Append gets a stream receiver, reads from it and records all the messages that are shipped to it
func (l CommandLoggerServer) Append(stream pb.CommandLogger_AppendServer) error {
	for {
		l, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if err := logs.Append(l.JobID, l.Line); err != nil {
			logrus.Errorf("Failed to record log entry %#v", l)
		}
	}
	return stream.SendAndClose(&pb.Empty{})
}

// CommandPipelineServer is used to send commands to remote executors
type CommandPipelineServer struct{}

// Select initializes a CommandPipelineServer
//
// It receives an AgentConfiguration which declares the commands that the remote
// executor is capable of running and a stream that will be used to send commands to
//
// It's not directly called, but using the remote client.
func (c CommandPipelineServer) Select(cfg *pb.AgentConfiguration, stream pb.CommandPipeline_SelectServer) error {
	logrus.Infof("Token: %s", cfg.Token)
	logrus.Infof("Labels: %s", cfg.Labels)
	logrus.Infof("Commands: %s", cfg.Commands)

	var jobID uint64
	for {
		jobID++
		err := stream.Send(&pb.CommandRequest{
			Command:     fmt.Sprintf("cmd-for-%s", cfg.Token),
			Args:        []string{"arg1", "arg2"},
			Channel:     "channel",
			ChannelID:   "channelID",
			ChannelLink: "channelLink",
			UserID:      "userID",
			Username:    "username",
			UserLink:    "userlink",
			JobID:       jobID,
			IsIM:        false,
		})
		if err == io.EOF {
			logrus.Info("The stream has been closed")
			return nil
		} else if err != nil {
			logrus.Errorf("Failed to send command %d to client: %s", jobID, err)
			return fmt.Errorf("something something")
		}
		select {
		case <-time.After(5 * time.Second):
			logrus.Debug("No data in over 5 seconds... looping.")
			continue
		case <-stream.Context().Done():
			logrus.Debug("bailing out, the context is done")
			return nil
		}
	}
	return nil
}

func (c CommandPipelineServer) Finish(_ context.Context, in *pb.CommandFinish) (*pb.Empty, error) {
	logrus.Infof("Changing job %d status to %s with error %s", in.JobID, in.Status, in.Error)
	return &pb.Empty{}, nil
}
