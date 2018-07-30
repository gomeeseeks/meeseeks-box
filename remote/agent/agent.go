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

	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// RemoteClient handles the configuration and the remote grpc client
type RemoteClient struct {
	config     Configuration
	grpcClient *grpc.ClientConn

	cmdClient api.CommandPipelineClient
	logClient api.LogWriterClient

	wg sync.WaitGroup

	ctx        context.Context
	cancelFunc context.CancelFunc

	agentID string
}

// New creates a new remote requester
func New(c Configuration) *RemoteClient {
	logrus.Debugf("creating new remote agent with configuration %#v", c)
	return &RemoteClient{
		agentID: uuid.NewV1().String(),
		config:  c,
		wg:      sync.WaitGroup{},
	}
}

// Connect creates a connection to the remote server
func (r *RemoteClient) Connect() error {
	logrus.Debugf("connecting to remote server: %s", r.config.ServerURL)

	ctx, cancel := context.WithTimeout(context.Background(), r.config.GetGRPCTimeout())
	defer cancel()

	c, err := grpc.DialContext(ctx, r.config.ServerURL, r.config.GetOptions())
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

	return r.config.registerLocalCommands()
}

// Run registers this agent in the remote server and launches a command stream to listen for commands to run
func (r *RemoteClient) Run() error {
	r.ctx, r.cancelFunc = context.WithCancel(context.Background())

	commandStream, err := r.cmdClient.RegisterAgent(r.ctx, r.config.createAgentConfiguration(r.agentID))
	if err != nil {
		return fmt.Errorf("failed to register commands on remote server: %s", err)
	}

	go r.run(commandStream) // TODO: when this method returns we should be quitting

	return nil
}

func (r *RemoteClient) run(pipeline api.CommandPipeline_RegisterAgentClient) {
	for {
		cmd, err := pipeline.Recv()
		if err == io.EOF {
			logrus.Fatalf("received EOF from command pipeline, quitting")
			r.triggerShutdown()
			return
		}
		if err != nil {
			s, ok := status.FromError(err)
			if !ok {
				logrus.Errorf("unknown error: %#v", err)
				continue
			}
			logrus.Errorf("grpc error, shutting down: %#v", *s)
			r.triggerShutdown()
			return
		}

		logrus.Debugf("received command from pipeline: %#v", cmd)
		r.wg.Add(1)
		go func(cmd api.CommandRequest) {
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
		}(*cmd)
	}
}

func (r *RemoteClient) triggerShutdown() {
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
