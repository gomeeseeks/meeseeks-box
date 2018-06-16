package client

import (
	"context"
	"fmt"
	"time"

	"github.com/gomeeseeks/meeseeks-box/config"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/metrics"
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

func (c *Configuration) newContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.GetTimeout())
}

func (c *Configuration) createAgentConfiguration() *api.AgentConfiguration {
	return &api.AgentConfiguration{
		Commands: c.createRemoteCommands(),
		Labels:   c.Labels,
		Token:    c.Token,
	}
}

// GetTimeout returns the configured timeout or a default of 10 seconds
func (c *Configuration) GetTimeout() time.Duration {
	if c.Timeout == 0 {
		return 10 * time.Second
	}
	return c.Timeout
}

func (c *Configuration) createRemoteCommands() map[string]*api.RemoteCommand {
	remoteCommands := make(map[string]*api.RemoteCommand, len(c.Commands))
	for name, cmd := range c.Commands {
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

// RemoteClient handles the configuration and the remote grpc client
type RemoteClient struct {
	config     Configuration
	grpcClient *grpc.ClientConn
}

// New creates a new remote requester
func New(c Configuration) *RemoteClient {
	return &RemoteClient{
		config: c,
		// requestsCh: make(chan meeseeks.Request, c.Pool),
	}
}

// Connect creates a connection to the remote server
func (r *RemoteClient) Connect() error {
	c, err := grpc.Dial(r.config.ServerURL, r.config.Options)
	if err != nil {
		return fmt.Errorf("could not connect to remote server %s: %s", r.config.ServerURL, err)
	}
	r.grpcClient = c
	return nil
}

// Requester creates a new requester using the client
func (r *RemoteClient) Requester() *RemoteRequester {
	return &RemoteRequester{
		client:     r,
		requestsCh: make(chan meeseeks.Request, r.config.Pool),
	}
}

// LogWriter creates a new LogWriter
func (r *RemoteClient) LogWriter() meeseeks.LogWriter {
	return grpcLogWriter{
		remoteClient: r,
	}
}

// LogReader returns a new remote log reader
func (r *RemoteClient) LogReader() meeseeks.LogReader {
	return nullReader{}
}

// RemoteRequester is an implementation of an executor.Listener
type RemoteRequester struct {
	client     *RemoteClient
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

// Start starts the client, listens to the remote channel and sends them to the execution pipeline
func (r *RemoteRequester) Start() error {
	if r.client.grpcClient == nil {
		return fmt.Errorf("client is not connected to a remote server")
	}

	cmdClient := api.NewCommandPipelineClient(r.client.grpcClient)

	ctx, cancel := r.client.config.newContext()
	defer cancel()

	// Register the commands and start listening on the stream of remote commands
	// On startup it has to send all the commands that the meeseeks knows how to handle (except builtins)
	commandStream, err := cmdClient.RegisterAgent(ctx, r.client.config.createAgentConfiguration())
	if err != nil {
		return fmt.Errorf("failed to register commands on remote server %s: %s", r.client.config.ServerURL, err)
	}

	// Get a command (or error)
	for {
		cmd, err := commandStream.Recv()
		if err != nil {
			logrus.Fatalf("failed to receive command: %s", err)
			break
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

	return nil
}

type grpcLogWriter struct {
	remoteClient *RemoteClient
}

// Append implements LogWritter.Append
func (g grpcLogWriter) Append(jobID uint64, content string) error {
	ctx, cancel := g.remoteClient.config.newContext()
	defer cancel()

	c := api.NewLogWriterClient(g.remoteClient.grpcClient)
	w, e := c.NewWriter(ctx)
	if e != nil {
		return e
	}

	metrics.LogLinesCount.Inc()
	return w.Send(&api.LogEntry{
		JobID: jobID,
		Line:  content,
	})
}

// SetError implements LogWritter.SetError
func (g grpcLogWriter) SetError(jobID uint64, jobErr error) error {
	ctx, cancel := g.remoteClient.config.newContext()
	defer cancel()

	c := api.NewLogWriterClient(g.remoteClient.grpcClient)
	_, err := c.SetError(ctx, &api.ErrorLogEntry{
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
