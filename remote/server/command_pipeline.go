package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gomeeseeks/meeseeks-box/commands"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/persistence"
	"github.com/gomeeseeks/meeseeks-box/remote/api"
	"github.com/sirupsen/logrus"
)

type finishPayload struct {
	err error
}

type commandPipelineServer struct {
	runningCommands map[uint64]chan interface{}
	agent           api.CommandPipeline_RegisterAgentServer

	wg   sync.WaitGroup
	lock sync.Mutex
}

func newCommandPipelineServer() *commandPipelineServer {
	return &commandPipelineServer{
		wg:              sync.WaitGroup{},
		lock:            sync.Mutex{},
		runningCommands: make(map[uint64]chan interface{}),
	}
}

// RegisterAgent registers a new agent service
func (p *commandPipelineServer) RegisterAgent(in *api.AgentConfiguration, agent api.CommandPipeline_RegisterAgentServer) error {
	// TODO: check the in.GetToken()
	// TODO: register the commands using the in.GetLabels()

	cmds := make([]commands.CommandRegistration, len(in.GetCommands()))
	for name, cmd := range in.Commands {
		cmds = append(cmds, commands.CommandRegistration{
			Name: name,
			Cmd: remoteCommand{
				CommandOpts: meeseeks.CommandOpts{
					Cmd:             name,
					AllowedChannels: cmd.GetAllowedChannels(),
					AllowedGroups:   cmd.GetAllowedGroups(),
					AuthStrategy:    cmd.GetAuthStrategy(),
					ChannelStrategy: cmd.GetChannelStrategy(),
					Templates:       cmd.GetTemplates(),
					Timeout:         time.Duration(cmd.GetTimeout()),
					Help: meeseeks.NewHelp(
						cmd.GetHelp().GetSummary(),
						cmd.GetHelp().GetArgs()...),
				},
				server: p,
			},
		})
	}

	if err := commands.Add(cmds...); err != nil {
		return fmt.Errorf("failed to register remote commands: %s", err)
	}
	p.agent = agent
	return nil
}

// Finish implements the finish server method
func (p *commandPipelineServer) Finish(ctx context.Context, fin *api.CommandFinish) (*api.Empty, error) {
	var err error
	if fin.GetError() != "" {
		err = errors.New(fin.GetError())
	}

	return &api.Empty{}, p.finishJob(fin.GetJobID(), err)
}

func (p *commandPipelineServer) appendJob(job meeseeks.Job) (chan interface{}, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	_, ok := p.runningCommands[job.ID]
	if ok {
		return nil, fmt.Errorf("Job %d is already in the list", job.ID)
	}

	req := job.Request
	err := p.agent.Send(&api.CommandRequest{
		Command: req.Command,
		Args:    req.Args,

		IsIM:        req.IsIM,
		Channel:     req.Channel,
		ChannelID:   req.ChannelID,
		ChannelLink: req.ChannelLink,
		UserID:      req.UserID,
		UserLink:    req.UserLink,
		Username:    req.Username,

		JobID: job.ID,
	})
	if err == io.EOF {
		return nil, fmt.Errorf("failed to send job %#v the remote stream is closed: %s", job, err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to send job %#v to remote executor: %s", job, err)
	}

	p.wg.Add(1)

	c := make(chan interface{})
	p.runningCommands[job.ID] = c

	return c, nil
}

func (p *commandPipelineServer) finishJob(jobID uint64, joberr error) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	c, ok := p.runningCommands[jobID]
	if !ok {
		return fmt.Errorf("Job %d is not registered as a running job", jobID)
	}

	defer p.wg.Done()

	if err := persistence.LogWriter().SetError(jobID, joberr); err != nil {
		logrus.Errorf("Failed to set error for job %d: %s", jobID, err)
	}

	c <- jobID

	return nil
}

type remoteCommand struct {
	meeseeks.CommandOpts
	server *commandPipelineServer
}

func (r remoteCommand) Execute(ctx context.Context, job meeseeks.Job) (string, error) {
	c, err := r.server.appendJob(job)
	if err != nil {
		return "", fmt.Errorf("failed to append remote job: %s", err)
	}

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("command failed because of context done: %s", ctx.Err())

	case <-c:
		jobLog, err := persistence.LogReader().Get(job.ID)
		if err != nil {
			return "", fmt.Errorf("failed to get the execution log from job %d", job.ID)
		}
		return jobLog.Output, jobLog.GetError()

	}

}
