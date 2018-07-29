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
	runningCommands map[uint64]chan finishedJob

	pipe chan api.CommandRequest

	wg   sync.WaitGroup
	lock sync.Mutex
}

func newCommandPipelineServer() *commandPipelineServer {
	return &commandPipelineServer{
		wg:              sync.WaitGroup{},
		lock:            sync.Mutex{},
		runningCommands: make(map[uint64]chan finishedJob),
		pipe:            make(chan api.CommandRequest),
	}
}

// RegisterAgent registers a new agent service
func (p *commandPipelineServer) RegisterAgent(in *api.AgentConfiguration, agent api.CommandPipeline_RegisterAgentServer) error {
	// TODO: check the in.GetToken()
	// TODO: register the commands using the in.GetLabels()

	p.registerCommands(in.Commands)

	for req := range p.pipe {
		err := agent.Send(&req)
		logrus.Debugf("request %#v sent to remote agent", req)

		if err == io.EOF {
			logrus.Warnf("remote agent is erring with EOF")
			continue
			// return nil //, fmt.Errorf("failed to send job %#v the remote stream is closed: %s", job, err)
		}
		if err != nil {
			logrus.Warnf("remote agent is erring with %s", err)
			return fmt.Errorf("failed to send job request %#v to remote executor: %s", req, err)
		}
	}

	return nil
}

func (p *commandPipelineServer) registerCommands(agentCommands map[string]*api.RemoteCommand) error {
	cmds := make([]commands.CommandRegistration, 0)
	for name, cmd := range agentCommands {
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
					Timeout:         time.Duration(cmd.GetTimeout()) * time.Second,
					Help: meeseeks.NewHelp(
						cmd.GetHelp().GetSummary(),
						cmd.GetHelp().GetArgs()...),
				},
				server: p,
			},
		})
	}

	logrus.Debugf("remote agent is registering commands %#v", cmds)
	if err := commands.Add(cmds...); err != nil {
		return fmt.Errorf("failed to register remote commands: %s", err)
	}
	return nil
}

func (p *commandPipelineServer) unregisterCommands(agentCommands map[string]*api.RemoteCommand) {
	cmds := make([]string, 0)
	for name := range agentCommands {
		cmds = append(cmds, name)
	}
	commands.Remove(cmds...)
}

// Finish implements the finish server method
func (p *commandPipelineServer) Finish(ctx context.Context, fin *api.CommandFinish) (*api.Empty, error) {
	logrus.Debugf("got %#v from remote agent", fin)
	return &api.Empty{}, p.finishJob(fin.GetJobID(), finishedJob{
		content: fin.GetContent(),
		err:     fin.GetError(),
	})
}

func (p *commandPipelineServer) appendJob(job meeseeks.Job) (chan finishedJob, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	_, ok := p.runningCommands[job.ID]
	if ok {
		return nil, fmt.Errorf("Job %d is already in the list", job.ID)
	}

	req := job.Request
	p.pipe <- api.CommandRequest{
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
	}

	p.wg.Add(1)
	c := make(chan finishedJob)
	p.runningCommands[job.ID] = c

	return c, nil
}

func (p *commandPipelineServer) finishJob(jobID uint64, f finishedJob) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	c, ok := p.runningCommands[jobID]
	if !ok {
		return fmt.Errorf("Job %d is not registered as a running job", jobID)
	}

	defer p.wg.Done()

	if f.getError() != nil {
		if err := persistence.LogWriter().SetError(jobID, f.getError()); err != nil {
			logrus.Errorf("Failed to set error for job %d: %s", jobID, err)
		}
	}

	logrus.Debugf("sending finished job %#v to channel", f)
	c <- f

	return nil
}

type remoteCommand struct {
	meeseeks.CommandOpts
	server *commandPipelineServer
}

func (r remoteCommand) Execute(ctx context.Context, job meeseeks.Job) (string, error) {
	logrus.Debug("start execution of job %#v", job)
	c, err := r.server.appendJob(job)
	if err != nil {
		return "", fmt.Errorf("failed to append remote job: %s", err)
	}

	select {
	case <-ctx.Done():
		logrus.Debug("job %#v failed with error %s", job, ctx.Err())
		return "", fmt.Errorf("command failed because of context done: %s", ctx.Err())

	case f := <-c:
		logrus.Debug("successful execution of job %#v with result %#v", job, f)
		return f.getContent(), f.getError()

	}
}

type finishedJob struct {
	content string
	err     string
}

func (f finishedJob) getContent() string {
	return f.content
}

func (f finishedJob) getError() error {
	if f.err != "" {
		return errors.New(f.err)
	}
	return nil
}
