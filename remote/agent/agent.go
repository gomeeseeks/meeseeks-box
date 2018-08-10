package agent

import (
	"context"
	"fmt"
	"io"
	"sync"
	"syscall"
	"time"

	"github.com/gomeeseeks/meeseeks-box/commands"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/persistence"
	"github.com/gomeeseeks/meeseeks-box/remote/api"

	"github.com/google/uuid"
	"github.com/jpillora/backoff"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RemoteClient handles the configuration and the remote grpc client
type RemoteClient struct {
	config     Configuration
	grpcClient *grpc.ClientConn

	cmdClient api.CommandPipelineClient
	logClient api.LogWriterClient

	pipeline api.CommandPipeline_RegisterAgentClient

	wg sync.WaitGroup

	ctx        context.Context
	cancelFunc context.CancelFunc

	agentID string
}

// New creates a new remote requester
func New(c Configuration) *RemoteClient {
	logrus.Debugf("creating new remote agent with configuration %#v", c)
	return &RemoteClient{
		agentID: uuid.New().String(),
		config:  c,
		wg:      sync.WaitGroup{},
	}
}

// Connect creates a connection to the remote server
func (r *RemoteClient) Connect() error {
	logrus.Debugf("connecting to remote server: %s", r.config.ServerURL)

	c, err := grpc.Dial(r.config.ServerURL, r.config.GetOptions()...)
	if err != nil {
		return fmt.Errorf("could not connect to remote server %s: %s", r.config.ServerURL, err)
	}

	logrus.Infof("connected to remote server: %s", r.config.ServerURL)
	r.cmdClient = api.NewCommandPipelineClient(c)
	r.logClient = api.NewLogWriterClient(c)
	r.grpcClient = c

	persistence.Register(
		persistence.Providers{
			LogReader: nullReader{},
			LogWriter: grpcLogWriter{
				client:         r.logClient,
				timeoutSeconds: r.config.GetGRPCTimeout(),
			},
		},
	)

	return nil
}

// Reconnect closes the connection pipeline
func (r *RemoteClient) Reconnect() {
	// This is not really handling the whole reconnection thing and can end up in
	// the whole thing dying, I need to test this properly
	r.cancelFunc()
}

// Run registers this agent in the remote server and launches a command stream to listen for commands to run
func (r *RemoteClient) Run() {

	b := &backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    1 * time.Second,
		Factor: 2,
		Jitter: true,
	}
	r.ctx, r.cancelFunc = context.WithCancel(context.Background())

Service:
	for {

		commandStream, err := r.cmdClient.RegisterAgent(r.ctx, r.config.createAgentConfiguration(r.agentID))
		if err != nil {
			if b.Attempt() > 10 {
				logrus.Errorf("failed to register agent in remote server: %s", err)
				r.triggerShutdown()
				return
			}
			logrus.Warnf("failed to register agent in remote server: %s... retrying", err)
			time.Sleep(b.Duration())
			continue Service
		}
		b.Reset()

		for {
			cmd, err := commandStream.Recv()
			if err == io.EOF {
				logrus.Infof("received EOF, shutting down")
				r.triggerShutdown()
				return
			}

			s := status.Code(err)
			switch s {
			case codes.OK:
				logrus.Debugf("all is good, continue")

			case codes.Unavailable:
				logrus.Infof("server is unavailable, reconnecting...")
				time.Sleep(time.Millisecond)
				continue Service

			case codes.Canceled:
				logrus.Infof("cancelled, quitting")
				r.triggerShutdown()
				return

			default:
				logrus.Errorf("grpc error %d, shutting down", s)
				r.triggerShutdown()
				return

			}

			logrus.Debugf("received command from pipeline: %#v", cmd)

			r.wg.Add(1)
			go r.runCommand(*cmd)
		}
	}
}

func (r *RemoteClient) runCommand(cmd api.CommandRequest) {
	defer r.wg.Done()

	// add a metric to account for remotely received commands
	rq := meeseeks.Request{
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

	logrus.Debugf("executing request: %#v", rq)
	localCmd, ok := commands.Find(&rq)
	if !ok {
		ctx, cancel := context.WithTimeout(r.ctx, r.config.GetGRPCTimeout())
		defer cancel()

		r.cmdClient.Finish(ctx, &api.CommandFinish{
			AgentID: r.agentID,
			JobID:   cmd.GetJobID(),
			Error:   fmt.Sprintf("could not find command %s in remote agent", cmd.GetCommand()),
		})
		return
	}

	logrus.Debugf("found command %#v", localCmd)
	ctx, cancelShellCmd := context.WithTimeout(r.ctx, localCmd.GetTimeout())
	defer cancelShellCmd()

	content, err := localCmd.Execute(ctx, meeseeks.Job{
		ID:        cmd.GetJobID(),
		Request:   rq,
		Status:    meeseeks.JobRunningStatus,
		StartTime: time.Now(),
	})

	var errString string
	if err != nil {
		errString = err.Error()
	}

	ctx, cancel := context.WithTimeout(r.ctx, r.config.GetGRPCTimeout())
	defer cancel()

	logrus.Debugf("sending command finish event %#v", cmd)
	r.cmdClient.Finish(ctx, &api.CommandFinish{
		AgentID: r.agentID,
		JobID:   cmd.GetJobID(),
		Content: content,
		Error:   errString,
	})
	logrus.Debugf("command %#v finished execution", cmd)
}

func (r *RemoteClient) triggerShutdown() {
	if r.pipeline != nil {
		if err := r.pipeline.CloseSend(); err != nil {
			logrus.Errorf("failed to send closing signal to server: %s", err)
		}
	}
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
}

// Shutdown will close the stream and wait for all the commands to finish execution
func (r *RemoteClient) Shutdown() {
	logrus.Debugf("invoking cancel function")
	r.cancelFunc()

	logrus.Debugf("waiting on sync wait group")
	r.wg.Wait()

	logrus.Debugf("done waiting, shutdown complete")
}
