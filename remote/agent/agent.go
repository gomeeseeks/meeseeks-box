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
	return &RemoteClient{
		config: c,
		wg:     sync.WaitGroup{},
	}
}

// Connect creates a connection to the remote server
func (r *RemoteClient) Connect() error {
	c, err := grpc.Dial(r.config.ServerURL, r.config.GetOptions())
	if err != nil {
		return fmt.Errorf("could not connect to remote server %s: %s", r.config.ServerURL, err)
	}

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

	go r.start(commandStream)

	return nil
}

func (r *RemoteClient) start(pipeline api.CommandPipeline_RegisterAgentClient) {
	for {
		cmd, err := pipeline.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			logrus.Errorf("error receiving command: %#v", err)
			return
		}

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

			ctx, cancelShellCmd := context.WithTimeout(context.Background(), localCmd.GetTimeout())
			defer cancelShellCmd()

			_, err = localCmd.Execute(ctx, meeseeks.Job{
				ID:        cmd.GetJobID(),
				Request:   rq,
				Status:    meeseeks.JobRunningStatus,
				StartTime: time.Now(),
			})

			ctx, cancel := context.WithTimeout(context.Background(), r.config.GetGRPCTimeout())
			defer cancel()

			r.cmdClient.Finish(ctx, &api.CommandFinish{
				JobID: cmd.GetJobID(),
				Error: err.Error(),
			})
		}(*cmd)
	}
}

// Shutdown will close the stream and wait for all the commands to finish execution
func (r *RemoteClient) Shutdown() {
	r.cancelFunc()
	r.wg.Wait()
}
