package agent

// To start the process in agent mode we need to register
//  Persistence.LogWriter (so we can write logs)
//  Persistence.Jobs (so we can finish processes)

import (
	"context"
	"fmt"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/remote/api"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// RemoteClient handles the configuration and the remote grpc client
type RemoteClient struct {
	config     Configuration
	grpcClient *grpc.ClientConn

	cmdClient api.CommandPipelineClient
	logClient api.LogWriterClient
}

// New creates a new remote requester
func New(c Configuration) *RemoteClient {
	return &RemoteClient{
		config: c,
	}
}

// Connect creates a connection to the remote server
func (r *RemoteClient) Connect() error {
	c, err := grpc.Dial(r.config.ServerURL, r.config.Options)
	if err != nil {
		return fmt.Errorf("could not connect to remote server %s: %s", r.config.ServerURL, err)
	}

	r.cmdClient = api.NewCommandPipelineClient(c)
	r.logClient = api.NewLogWriterClient(c)
	r.grpcClient = c

	return nil
}

// RemoteLogReader returns a new remote log reader
func (r *RemoteClient) RemoteLogReader() meeseeks.LogReader {
	return nullReader{}
}

// RemoteLogWriter creates a new LogWriter
func (r *RemoteClient) RemoteLogWriter() meeseeks.LogWriter {
	return grpcLogWriter{
		client:         r.logClient,
		timeoutSeconds: r.config.GetTimeout(),
	}
}

// RemoteJobs creates a new Jobs object that implements the meeseeks.Jobs interface
func (r *RemoteClient) RemoteJobs() meeseeks.Jobs {
	return remoteJobs{
		client:         r.cmdClient,
		timeoutSeconds: r.config.GetTimeout(),
	}
}

// CreateRequester creates a new requester, registering the agent and starting it so it is ready to take remote requests
func (r *RemoteClient) CreateRequester() (*RemoteRequester, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.config.GetTimeout())
	defer cancel()

	// Register the commands and start listening on the stream of remote commands
	// On startup it has to send all the commands that the meeseeks knows how to handle (except builtins)
	commandStream, err := r.cmdClient.RegisterAgent(ctx, r.config.createAgentConfiguration())
	if err != nil {
		return nil, fmt.Errorf("failed to register commands on remote server: %s", err)
	}

	requester := &RemoteRequester{
		pipeline:   commandStream,
		requestsCh: make(chan meeseeks.Request, r.config.Pool),
	}

	go requester.start()

	return requester, nil
}

// RemoteRequester is an implementation of an executor.Listener
type RemoteRequester struct {
	pipeline api.CommandPipeline_RegisterAgentClient

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
func (r *RemoteRequester) start() {
	for {
		cmd, err := r.pipeline.Recv()
		if err != nil {
			logrus.Errorf("failed to receive command: %#v", err)
			return
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
}
