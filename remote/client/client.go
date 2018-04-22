package client

import (
	"context"
	"fmt"
	"time"

	"github.com/gomeeseeks/meeseeks-box/config"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/remote/api"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Requester interface {
	RequestsCh() chan meeseeks.Request
}

type RemoteRequester struct {
	config     Configuration
	grpcClient *grpc.ClientConn
	requestCh  chan meeseeks.Request
}

func New(c Configuration) RemoteRequester {
	return RemoteRequester{
		config:    c,
		requestCh: make(chan meeseeks.Request, c.Pool),
	}
}

func (this RemoteRequester) Connect() error {
	c, err := grpc.Dial(this.config.ServerURL, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("could not connect to remote server %s: %s", this.config.ServerURL, err)
	}
	this.grpcClient = c
	return nil
}

func (this RemoteRequester) RequestsCh() chan meeseeks.Request {
	return this.requestCh
}

type Configuration struct {
	ServerURL string
	Timeout   time.Duration
	Token     string
	Pool      int
	Labels    map[string]string
	Commands  map[string]config.Command
}

func (this Configuration) createRemoteCommands() map[string]*api.RemoteCommand {
	remoteCommands := make(map[string]*api.RemoteCommand, len(this.Commands))
	for name, cmd := range this.Commands {
		remoteCommands[name] = &api.RemoteCommand{
			Timeout:         int64(cmd.Timeout),
			AuthStrategy:    cmd.AuthStrategy,
			AllowedGroups:   cmd.AllowedGroups,
			ChannelStrategy: cmd.ChannelStrategy,
			AllowedChannels: cmd.AllowedChannels,
			Templates:       cmd.Templates,
			Help: &api.Help{
				Summary: cmd.Help.Summary,
				Args:    cmd.Help.Args,
			},
		}
	}
	return remoteCommands
}

func (this Configuration) createAgentConfiguration() *api.AgentConfiguration {
	return &api.AgentConfiguration{
		Commands: this.createRemoteCommands(),
		Labels:   this.Labels,
		Token:    this.Token,
	}
}

func (this Configuration) newContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), this.GetTimeout())
}

func (this Configuration) GetTimeout() time.Duration {
	if this.Timeout == 0 {
		return 10 * time.Second
	}
	return this.Timeout
}

func (this RemoteRequester) Listen() error {
	if this.grpcClient == nil {
		return fmt.Errorf("client is not connected to a remote server")
	}

	cmdClient := api.NewCommandPipelineClient(this.grpcClient)

	ctx, cancel := this.config.newContext()
	defer cancel()

	// Register the commands and start listening on the stream of remote commands
	// On startup it has to send all the commands that the meeseeks knows how to handle (except builtins)
	stream, err := cmdClient.RegisterAgent(ctx, this.config.createAgentConfiguration())
	if err != nil {
		return fmt.Errorf("failed to register commands on remote server %s: %s", this.config.ServerURL, err)
	}

	// The whole logger thing should happen on connection by registering the logger for the shell commands
	// // Get a logAppender
	// logCtx, logCnl := this.config.newContext()
	// defer logCnl()

	// loggerClient := api.NewCommandLoggerClient(this.grpcClient)
	// appender, err := loggerClient.NewAppender(logCtx)
	// if err != nil {
	// 	// Add metric for failing to get a log appender
	// 	logrus.Fatalf("Could not open an appender stream: %s", err)
	// }

	// Get a command (or error)
	for {
		cmd, err := stream.Recv()
		if err != nil {
			logrus.Fatalf("failed to receive a command: %s", err)
		}
		// add a metric to account for remotely received commands
		this.requestCh <- meeseeks.Request{
			Command:     cmd.Command,
			Args:        cmd.Args,
			Channel:     cmd.Channel,
			ChannelID:   cmd.ChannelID,
			ChannelLink: cmd.ChannelLink,
			IsIM:        cmd.IsIM,
			UserID:      cmd.UserID,
			Username:    cmd.Username,
			UserLink:    cmd.UserLink,
		}
	}

	// for each command
	// Execute the command and send all the logs
	// for _, line := range []string{"line1", "line2", "line3", "line4"} {
	// 	l := api.LogEntry{
	// 		JobID: cmd.JobID,
	// 		Line:  line}
	// 	if err := appender.Send(&l); err != nil {
	// 		logrus.Fatalf("Failed to send log line to backend: %s", err)
	// 	}
	// 	// add metric for sent log lines
	// }
	// if _, err := appender.CloseAndRecv(); err != nil {
	// 	logrus.Fatalf("failed to close the channel: %s", err)
	// }
	// // Add metric for remotely finished commands
	// _, err = commands.Finish(context.Background(), &api.CommandFinish{
	// 	JobID:  cmd.JobID,
	// 	Status: "Done",
	// 	Error:  "",
	// })
	// if err != nil {
	// 	logrus.Fatalf("Failed to finish the command execution %s", err)
	// }

	return nil
}
