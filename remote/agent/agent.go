package agent

// To start the process in agent mode we need to register
//  Persistence.LogWriter (so we can write logs)
//  Persistence.Jobs (so we can finish processes)

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gomeeseeks/meeseeks-box/commands"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/persistence"
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

	wg         sync.WaitGroup
	cancelFunc context.CancelFunc
}

// New creates a new remote requester
func New(c Configuration) *RemoteClient {
	logrus.Debugf("creating new remote agent with configuration %#v", c)
	return &RemoteClient{
		config: c,
		wg:     sync.WaitGroup{},
	}
}

// Connect creates a connection to the remote server
func (r *RemoteClient) Connect() error {
	logrus.Debugf("connecting to remote server: %s", r.config.ServerURL)
	c, err := grpc.Dial(r.config.ServerURL, r.config.GetOptions())
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

// RegisterAndRun creates a new requester, registering the agent and starting it so it is ready to take remote requests
func (r *RemoteClient) RegisterAndRun() error {
	ctx, cancel := context.WithCancel(context.Background())
	r.cancelFunc = cancel

	commandStream, err := r.cmdClient.RegisterAgent(ctx, r.config.createAgentConfiguration())
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
			logrus.Debugf("received EOF from command pipeline, quitting")
			return
		}
		if err != nil {
			logrus.Errorf("error receiving command: %#v", err)
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
				ctx, cancel := context.WithTimeout(context.Background(), r.config.GetGRPCTimeout())
				defer cancel()

				r.cmdClient.Finish(ctx, &api.CommandFinish{
					JobID: cmd.GetJobID(),
					Error: fmt.Sprintf("could not find command %s in remote agent", cmd.GetCommand()),
				})
				return
			}

			logrus.Debugf("found command %#v", localCmd)
			ctx, cancelShellCmd := context.WithTimeout(context.Background(), localCmd.GetTimeout())
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

			ctx, cancel := context.WithTimeout(context.Background(), r.config.GetGRPCTimeout())
			defer cancel()

			logrus.Debugf("sending command finish event %#v", cmd)
			r.cmdClient.Finish(ctx, &api.CommandFinish{
				JobID:   cmd.GetJobID(),
				Content: content,
				Error:   errString,
			})
			logrus.Debugf("command %#v finished execution", cmd)
		}(*cmd)
	}
}

// Shutdown will close the stream and wait for all the commands to finish execution
func (r *RemoteClient) Shutdown() {
	r.cancelFunc()
	r.wg.Wait()
}
