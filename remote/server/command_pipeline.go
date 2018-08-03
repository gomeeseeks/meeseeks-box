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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type finishPayload struct {
	err error
}

type commandPipelineServer struct {
	runningJobs map[uint64]chan finishedJob

	lock *sync.Mutex
}

func newCommandPipelineServer() *commandPipelineServer {
	return &commandPipelineServer{
		runningJobs: make(map[uint64]chan finishedJob),

		lock: &sync.Mutex{},
	}
}

type jobStarter interface {
	StartJob(req api.CommandRequest) chan finishedJob
}

// RegisterAgent registers a new agent service
func (p *commandPipelineServer) RegisterAgent(in *api.AgentConfiguration, agent api.CommandPipeline_RegisterAgentServer) error {
	// TODO: check the in.GetToken()
	// TODO: register the commands using the in.GetLabels()

	pipe, err := p.registerAgent(in)
	if err != nil {
		return fmt.Errorf("failed to register remote agent %s: %s", in.GetAgentID(), err)
	}

	go func() {
		select {
		case <-agent.Context().Done():
			logrus.Infof("agent %s context is done in server with error %s, closing pipe", in.GetAgentID(), agent.Context().Err())
			close(pipe)
		}
	}()

Loop:
	for req := range pipe {
		err := agent.Send(&req)
		logrus.Debugf("request %#v sent to remote agent %s", req, in.GetAgentID())

		if err == io.EOF {
			logrus.Infof("remote agent %s is erring with EOF, quitting", in.GetAgentID())
			p.finishJob(finishedJob{
				jobID:   req.GetJobID(),
				agentID: in.GetAgentID(),
				content: "",
				err:     fmt.Sprintf("remote agent %s erred out with EOF, it seems to be gone", in.GetAgentID()),
			})
			close(pipe)
			break Loop
		}

		errCode := status.Code(err)
		switch errCode {
		case codes.OK:
			logrus.Debugf("agent %s received the job OK, continuing", in.GetAgentID())
			continue

		case codes.Canceled, codes.DeadlineExceeded:
			logrus.Infof("agent %s cancelled or had a timeout, it seems to be gone: %v - %s", in.GetAgentID(), errCode, err)

		default:
			logrus.Errorf("agent %s erred out with: %v - %s", in.GetAgentID(), errCode, err)

		}
		close(pipe)
		break Loop
	}

	logrus.Infof("unregistering remote agent %s", in.GetAgentID())
	p.deRegisterAgentCommands(in)

	return nil
}

// Finish implements the finish server method
func (p *commandPipelineServer) Finish(ctx context.Context, fin *api.CommandFinish) (*api.Empty, error) {
	logrus.Debugf("got %#v from remote agent", fin)
	return &api.Empty{}, p.finishJob(finishedJob{
		agentID: fin.GetAgentID(),
		jobID:   fin.GetJobID(),
		content: fin.GetContent(),
		err:     fin.GetError(),
	})
}

func (p *commandPipelineServer) registerAgent(in *api.AgentConfiguration) (chan api.CommandRequest, error) {

	agentPipe := make(chan api.CommandRequest)

	agent := remoteAgent{
		agentID:   in.GetAgentID(),
		agentPipe: agentPipe,

		jobStarter: p,
	}

	cmds := make([]commands.CommandRegistration, 0)
	for name, cmd := range in.Commands {
		cmds = append(cmds, commands.CommandRegistration{
			Name: name,
			Cmd: remoteCommand{
				agent: agent,
				CommandOpts: meeseeks.CommandOpts{
					Cmd:             name,
					AllowedChannels: cmd.GetAllowedChannels(),
					AllowedGroups:   cmd.GetAllowedGroups(),
					AuthStrategy:    cmd.GetAuthStrategy(),
					ChannelStrategy: cmd.GetChannelStrategy(),
					Handshake:       cmd.GetHasHandshake(),
					Timeout:         time.Duration(cmd.GetTimeout()) * time.Second,
					Help: meeseeks.NewHelp(
						cmd.GetHelp().GetSummary(),
						cmd.GetHelp().GetArgs()...),
				},
			},
		})
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	logrus.Debugf("remote agent is registering commands %#v", cmds)
	if err := commands.Add(cmds...); err != nil {
		return nil, fmt.Errorf("failed to register remote commands: %s", err)
	}

	return agentPipe, nil
}

func (p *commandPipelineServer) deRegisterAgentCommands(in *api.AgentConfiguration) {

	cmds := make([]string, 0)
	for name := range in.Commands {
		cmds = append(cmds, name)
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	commands.Remove(cmds...)
}

func (p *commandPipelineServer) finishJob(f finishedJob) error {
	c, err := p.PopJob(f.jobID)
	if err != nil {
		return fmt.Errorf("could not fetch command finish channel: %s", err)
	}

	if f.getError() != nil {
		if err := persistence.LogWriter().SetError(f.jobID, f.getError()); err != nil {
			logrus.Errorf("Failed to set error for job %d: %s", f.jobID, err)
		}
	}

	logrus.Debugf("sending finished job signal %#v to channel", f)

	c <- f

	return nil
}

func (p *commandPipelineServer) StartJob(req api.CommandRequest) chan finishedJob {
	p.lock.Lock()
	defer p.lock.Unlock()

	c := make(chan finishedJob)
	p.runningJobs[req.GetJobID()] = c
	return c
}

func (p *commandPipelineServer) PopJob(jobID uint64) (chan finishedJob, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	c, ok := p.runningJobs[jobID]
	if !ok {
		return nil, fmt.Errorf("could not find job with ID %d in the running jobs list", jobID)
	}

	delete(p.runningJobs, jobID)

	return c, nil
}

type remoteAgent struct {
	agentID string

	agentPipe chan api.CommandRequest

	jobStarter
}

func (r *remoteAgent) start(req api.CommandRequest) chan finishedJob {
	c := r.StartJob(req)
	r.agentPipe <- req

	return c
}

type remoteCommand struct {
	meeseeks.CommandOpts

	agent remoteAgent
}

func (r remoteCommand) Execute(ctx context.Context, job meeseeks.Job) (string, error) {
	logrus.Debug("start execution of job %#v", job)

	req := job.Request
	c := r.agent.start(api.CommandRequest{
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

	logrus.Debug("waiting for remote request to finish %#v", req)

	select {
	case <-ctx.Done():
		logrus.Debug("job %#v failed with error %s", job, ctx.Err())
		return "", fmt.Errorf("command failed because of context done: %s", ctx.Err())

	case f := <-c:
		logrus.Debug("successful execution of job %#v with result %#v", job, f)
		// TODO: check that the agent that finished the command is the same that started it
		return f.getContent(), f.getError()

	}
}

type finishedJob struct {
	agentID string
	jobID   uint64
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
