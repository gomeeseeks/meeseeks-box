package executor

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/gomeeseeks/meeseeks-box/auth"
	"github.com/gomeeseeks/meeseeks-box/commands"
	"github.com/gomeeseeks/meeseeks-box/commands/builtins"
	"github.com/gomeeseeks/meeseeks-box/formatter"
	"github.com/gomeeseeks/meeseeks-box/jobs"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/metrics"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/request"
	"github.com/gomeeseeks/meeseeks-box/messenger"
)

// ChatClient interface that provides a way of replying to messages on a channel
type ChatClient interface {
	Reply(formatter.Reply)
}

// Meeseeks is the command execution engine
type Meeseeks struct {
	client    ChatClient
	messenger *messenger.Messenger

	tasksCh        chan task
	wg             sync.WaitGroup
	activeCommands *activeCommands
}

type task struct {
	job meeseeks.Job
	cmd meeseeks.Command
}

// New creates a new Meeseeks service
func New(client ChatClient, messenger *messenger.Messenger) *Meeseeks {
	ac := newActiveCommands()
	commands.Add(builtins.BuiltinCancelJobCommand, builtins.NewCancelJobCommand(ac.Cancel))
	commands.Add(builtins.BuiltinKillJobCommand, builtins.NewKillJobCommand(ac.Cancel))

	m := Meeseeks{
		messenger: messenger,
		client:    client,
		tasksCh:   make(chan task, 20),

		wg:             sync.WaitGroup{},
		activeCommands: ac,
	}

	go m.loop()

	return &m
}

// Start launches the meeseeks to read messages from the MessageCh
func (m *Meeseeks) Start() {
	for msg := range m.messenger.MessagesCh() {
		req, err := request.FromMessage(msg)
		if err != nil {
			logrus.Debugf("Failed to parse message '%s' as a command: %s", msg.GetText(), err)
			m.client.Reply(formatter.Get().FailureReply(req, err))
			continue
		}
		metrics.ReceivedCommandsCount.Inc()

		cmd, ok := commands.Find(&req)
		if !ok {
			m.client.Reply(formatter.Get().UnknownCommandReply(req))
			metrics.UnknownCommandsCount.Inc()
			continue
		}

		if err = auth.Check(req, cmd); err != nil {
			m.client.Reply(formatter.Get().UnauthorizedCommandReply(req))
			metrics.RejectedCommandsCount.WithLabelValues(req.Command).Inc()
			continue
		}

		logrus.Infof("Accepted command '%s' from user '%s' on channel '%s' with args: %s",
			req.Command, req.Username, req.Channel, req.Args)
		metrics.AcceptedCommandsCount.WithLabelValues(req.Command).Inc()

		t, err := m.createTask(req, cmd)
		if err != nil {
			m.client.Reply(formatter.Get().FailureReply(req, fmt.Errorf("could not create task: %s", err)))
			continue
		}

		m.wg.Add(1)
		m.tasksCh <- t
	}
}

func (m *Meeseeks) createTask(req meeseeks.Request, cmd meeseeks.Command) (task, error) {
	if !cmd.Record() {
		return task{job: jobs.NullJob(req), cmd: cmd}, nil
	}

	j, err := jobs.Create(req)
	return task{job: j, cmd: cmd}, err
}

// Shutdown initiates a shutdown process by waiting for jobs to finish and then
// closing the tasks channel
func (m *Meeseeks) Shutdown() {
	defer m.closeTasksChannel()

	logrus.Info("Waiting for jobs to finish")
	m.wg.Wait()
	logrus.Info("Done waiting, exiting")
}

func (m *Meeseeks) closeTasksChannel() {
	logrus.Infof("Closing meeseeks tasks channel")
	close(m.tasksCh)
}

func (m *Meeseeks) loop() {
	for t := range m.tasksCh {
		go func(t task) {
			job := t.job
			req := job.Request
			cmd := t.cmd

			if cmd.HasHandshake() {
				m.client.Reply(formatter.Get().HandshakeReply(req))
			}

			ctx := m.activeCommands.Add(t)
			defer m.activeCommands.Cancel(job.ID)

			out, err := t.cmd.Execute(ctx, t.job)
			if err != nil {
				logrus.Errorf("Command '%s' from user '%s' failed execution with error: %s",
					req.Command, req.Username, err)

				m.client.Reply(formatter.Get().FailureReply(req, err).WithOutput(out))

				metrics.FailedTasksCount.WithLabelValues(job.Request.Command).Inc()
				jobs.Finish(job.ID, jobs.FailedStatus)

			} else {
				logrus.Infof("Command '%s' from user '%s' succeeded execution", req.Command,
					req.Username)

				m.client.Reply(formatter.Get().SuccessReply(req).WithOutput(out))

				metrics.SuccessfulTasksCount.WithLabelValues(job.Request.Command).Inc()
				jobs.Finish(job.ID, jobs.SuccessStatus)

			}
			m.wg.Done()
		}(t)
	}
}

type activeCommands struct {
	ctx map[uint64]context.CancelFunc
	m   sync.Mutex
}

func newActiveCommands() *activeCommands {
	return &activeCommands{
		ctx: make(map[uint64]context.CancelFunc),
	}
}

func (a *activeCommands) Add(t task) context.Context {
	defer a.m.Unlock()
	a.m.Lock()

	ctx, cancel := context.WithCancel(context.Background())
	a.ctx[t.job.ID] = cancel
	return ctx
}

func (a *activeCommands) Cancel(jobID uint64) {
	defer a.m.Unlock()
	a.m.Lock()

	cancel, ok := a.ctx[jobID]
	if !ok {
		logrus.Debugf("could not cancel job %d because it is not in the active jobs list", jobID)
		return
	}

	// Delete the cancel command from the map
	delete(a.ctx, jobID)

	// Invoke the cancel function
	cancel()
}
