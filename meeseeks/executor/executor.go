package executor

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"gitlab.com/yakshaving.art/meeseeks-box/auth"
	"gitlab.com/yakshaving.art/meeseeks-box/commands"
	"gitlab.com/yakshaving.art/meeseeks-box/commands/builtins"
	"gitlab.com/yakshaving.art/meeseeks-box/meeseeks"
	"gitlab.com/yakshaving.art/meeseeks-box/meeseeks/metrics"
	"gitlab.com/yakshaving.art/meeseeks-box/persistence"
	"gitlab.com/yakshaving.art/meeseeks-box/text/formatter"
)

// ChatClient interface that provides a way of replying to messages on a channel
type ChatClient interface {
	Reply(formatter.Reply)
}

// NullChatClient is a client that implements the interface but does nothing
type NullChatClient struct{}

// Reply implements ChatClient.Reply
func (NullChatClient) Reply(_ formatter.Reply) {}

// Listener provides the necessary interface to start listening requests from a channel.
type Listener interface {
	Listen(chan<- meeseeks.Request)
}

// Executor is the command execution engine
type Executor struct {
	client    ChatClient
	listeners []Listener

	requestsCh chan meeseeks.Request

	tasksCh        chan task
	wg             sync.WaitGroup
	activeCommands *activeCommands
}

type task struct {
	job meeseeks.Job
	cmd meeseeks.Command
}

// Args is handy to set multiple arguments
type Args struct {
	ConcurrentTaskCount int
	WithBuiltinCommands bool
	ChatClient          ChatClient
}

// New creates a new Meeseeks service
func New(args Args) *Executor {
	ac := newActiveCommands()
	if args.WithBuiltinCommands {
		builtins.LoadBuiltins(
			builtins.NewCancelJobCommand(ac.Cancel),
			builtins.NewKillJobCommand(ac.Cancel),
		)
	}

	e := Executor{
		client:     args.ChatClient,
		tasksCh:    make(chan task, args.ConcurrentTaskCount),
		requestsCh: make(chan meeseeks.Request),

		wg:             sync.WaitGroup{},
		activeCommands: ac,
	}

	go e.processTasks()

	return &e
}

// ListenTo appends a listener to the list and starts listening to it
func (m *Executor) ListenTo(l Listener) {
	logrus.Debugf("Executor: adding listener %#v", l)
	m.listeners = append(m.listeners, l)
	go l.Listen(m.requestsCh)
}

// Run launches the meeseeks to read requests from the requests channel
func (m *Executor) Run() {
	for req := range m.requestsCh {
		metrics.ReceivedCommandsCount.Inc()

		cmd, ok := commands.Find(&req)
		if !ok {
			m.client.Reply(formatter.UnknownCommandReply(req))
			metrics.UnknownCommandsCount.Inc()
			continue
		}

		if err := auth.Check(req, cmd); err != nil {
			m.client.Reply(formatter.UnauthorizedCommandReply(req))
			metrics.RejectedCommandsCount.WithLabelValues(req.Command).Inc()
			continue
		}

		logrus.Infof("Accepted command '%s' from user '%s' on channel '%s' with args: %s",
			req.Command, req.Username, req.Channel, req.Args)
		metrics.AcceptedCommandsCount.WithLabelValues(req.Command).Inc()

		t, err := m.createTask(req, cmd)
		if err != nil {
			m.client.Reply(formatter.FailureReply(req, fmt.Errorf("could not create task: %s", err)))
			continue
		}

		m.wg.Add(1)
		m.tasksCh <- t
	}
}

func (m *Executor) createTask(req meeseeks.Request, cmd meeseeks.Command) (task, error) {
	if !cmd.MustRecord() {
		return task{job: persistence.Jobs().Null(req), cmd: cmd}, nil
	}

	j, err := persistence.Jobs().Create(req)
	return task{job: j, cmd: cmd}, err
}

// Shutdown initiates a shutdown process by waiting for jobs to finish and then
// closing the tasks channel
func (m *Executor) Shutdown() {
	defer m.closeTasksChannel()

	logrus.Info("Waiting for jobs to finish")
	m.wg.Wait()
	logrus.Info("Done waiting, exiting")
}

func (m *Executor) closeTasksChannel() {
	logrus.Infof("Closing meeseeks tasks channel")
	close(m.tasksCh)
}

func (m *Executor) processTasks() {
	for t := range m.tasksCh {
		go func(t task) {
			job := t.job
			req := job.Request
			cmd := t.cmd

			if cmd.HasHandshake() {
				m.client.Reply(formatter.HandshakeReply(req))
			}

			ctx := m.activeCommands.Add(t)
			defer m.activeCommands.Cancel(job.ID)

			out, err := t.cmd.Execute(ctx, t.job)
			if err != nil {
				logrus.Errorf("Command '%s' from user '%s' failed execution with error: %s",
					req.Command, req.Username, err)

				m.client.Reply(formatter.FailureReply(req, err).WithOutput(out))

				persistence.Jobs().Fail(job.ID)

			} else {
				logrus.Infof("Command '%s' from user '%s' succeeded execution", req.Command,
					req.Username)

				m.client.Reply(formatter.SuccessReply(req).WithOutput(out))

				persistence.Jobs().Succeed(job.ID)
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
