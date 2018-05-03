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

// Configuration holds the client configuration used to connect to the server
type Configuration struct {
	ServerURL string
	Options   grpc.DialOption // grpc.WithInsecure()
	Timeout   time.Duration
	Token     string
	Pool      int
	Labels    map[string]string
	Commands  map[string]config.Command
}

// New creates a new remote requester
func New(c Configuration) *RemoteRequester {
	return &RemoteRequester{
		config:     c,
		requestsCh: make(chan meeseeks.Request, c.Pool),
	}
}

// RemoteRequester is an implementation of an executor.Listener
type RemoteRequester struct {
	config     Configuration
	grpcClient *grpc.ClientConn
	requestsCh chan meeseeks.Request
	stopCh     chan bool
}

// Listen implements Requester.Listen
func (r *RemoteRequester) Listen(ch chan<- meeseeks.Request) {
	for keepRunning := true; keepRunning; {
		select {
		case rq := <-r.requestsCh:
			ch <- rq
		case keepRunning = <-r.stopCh:
			// do nothing
		}
	}
}

// Connect creates a connection to the remote server
func (r *RemoteRequester) Connect() error {
	c, err := grpc.Dial(r.config.ServerURL, r.config.Options)
	if err != nil {
		return fmt.Errorf("could not connect to remote server %s: %s", r.config.ServerURL, err)
	}
	r.grpcClient = c
	return nil
}

// Start starts the client, listens to the remote channel and sends them to the execution pipeline
func (r *RemoteRequester) Start() error {
	if r.grpcClient == nil {
		return fmt.Errorf("client is not connected to a remote server")
	}

	cmdClient := api.NewCommandPipelineClient(r.grpcClient)

	ctx, cancel := r.config.newContext()
	defer cancel()

	// Register the commands and start listening on the stream of remote commands
	// On startup it has to send all the commands that the meeseeks knows how to handle (except builtins)
	stream, err := cmdClient.RegisterAgent(ctx, r.config.createAgentConfiguration())
	if err != nil {
		return fmt.Errorf("failed to register commands on remote server %s: %s", r.config.ServerURL, err)
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
		r.requestsCh <- meeseeks.Request{
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

func (r Configuration) createRemoteCommands() map[string]*api.RemoteCommand {
	remoteCommands := make(map[string]*api.RemoteCommand, len(r.Commands))
	for name, cmd := range r.Commands {
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

func (r Configuration) createAgentConfiguration() *api.AgentConfiguration {
	return &api.AgentConfiguration{
		Commands: r.createRemoteCommands(),
		Labels:   r.Labels,
		Token:    r.Token,
	}
}

func (r Configuration) newContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), r.GetTimeout())
}

// GetTimeout returns the configured timeout or a default of 10 seconds
func (r Configuration) GetTimeout() time.Duration {
	if r.Timeout == 0 {
		return 10 * time.Second
	}
	return r.Timeout
}

// LoggerProvider wraps the specific logger implementation
// type LoggerProvider interface {
// 	Reader(jobID uint64) LogReader
// 	Writer(jobID uint64) LogWriter
// }
type grpcLoggerProvider struct {
	conn *grpc.ClientConn
}

func newLoggerProvider(conn *grpc.ClientConn) grpcLoggerProvider {
	return grpcLoggerProvider{
		conn: conn,
	}
}

// Reader implements
func (g grpcLoggerProvider) Reader(jobID uint64) meeseeks.LogReader {
	return nullReader{jobID}
}

type nullReader struct {
	jobID uint64
}

// Get implements meeseeks.LogReader.Get
func (n nullReader) Get() (meeseeks.JobLog, error) {
	return meeseeks.JobLog{}, nil
}

// Head implements meeseeks.LogReader.Head
func (n nullReader) Head(int) (meeseeks.JobLog, error) {
	return meeseeks.JobLog{}, nil
}

// Tail implements meeseeks.LogReader.Tail
func (n nullReader) Tail(int) (meeseeks.JobLog, error) {
	return meeseeks.JobLog{}, nil
}

// Writer returns a LogWriter
func (g grpcLoggerProvider) Writer(jobID uint64) meeseeks.LogWriter {
	client := api.NewLogWriterClient(g.conn)
	appender, err := client.NewWriter(context.Background()) // TODO improve this with some options
	if err != nil {
		logrus.Fatalf("could not create appender: %s", err)
	}

	return grpcLogWriter{
		jobID:    jobID,
		client:   client,
		appender: appender,
	}
}

type grpcLogWriter struct {
	jobID    uint64
	client   api.LogWriterClient
	appender api.LogWriter_NewWriterClient
}

// Append implements LogWritter.Append
func (g grpcLogWriter) Append(content string) error {
	return g.appender.Send(&api.LogEntry{
		JobID: g.jobID,
		Line:  content,
	})
}

// SetError implements LogWritter.SetError
func (g grpcLogWriter) SetError(jobErr error) error {
	_, err := g.client.SetError(context.Background(), &api.ErrorLogEntry{
		JobID: g.jobID,
		Error: jobErr.Error(),
	})
	return err
}
