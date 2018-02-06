package meeseeks

import (
	"fmt"
	"sync"

	"github.com/prometheus/common/log"
	"github.com/sirupsen/logrus"

	"github.com/pcarranza/meeseeks-box/command"
	"github.com/pcarranza/meeseeks-box/formatter"
	"github.com/pcarranza/meeseeks-box/jobs"
	"github.com/pcarranza/meeseeks-box/messenger"

	"github.com/pcarranza/meeseeks-box/auth"
	"github.com/pcarranza/meeseeks-box/commands"
	"github.com/pcarranza/meeseeks-box/meeseeks/request"
)

// ChatClient interface that provides a way of replying to messages on a channel
type ChatClient interface {
	Reply(text, color, channel string) error
	ReplyIM(text, color, user string) error
}

// Meeseeks is the command execution engine
type Meeseeks struct {
	client    ChatClient
	messenger *messenger.Messenger
	formatter *formatter.Formatter

	tasksCh chan task
	wg      sync.WaitGroup
}

type task struct {
	job jobs.Job
	cmd command.Command
}

// New creates a new Meeseeks service
func New(client ChatClient, messenger *messenger.Messenger, formatter *formatter.Formatter) *Meeseeks {
	m := Meeseeks{
		messenger: messenger,
		formatter: formatter,
		client:    client,
		tasksCh:   make(chan task, 20),

		wg: sync.WaitGroup{},
	}

	go m.jobsLoop()

	return &m
}

// Start launches the meeseeks to read messages from the MessageCh
func (m *Meeseeks) Start() {
	for msg := range m.messenger.MessagesCh() {
		req, err := request.FromMessage(msg)
		if err != nil {
			logrus.Debugf("Failed to parse message '%s' as a command: %s", msg.GetText(), err)
			m.replyWithError(msg, err)
			continue
		}

		cmd, ok := commands.Find(req.Command)
		if !ok {
			m.replyWithUnknownCommand(req)
			continue
		}
		if err = auth.Check(req.Username, cmd); err != nil {
			m.replyWithUnauthorizedCommand(req, cmd)
			continue
		}

		logrus.Infof("Accepted command '%s' from user '%s' on channel '%s' with args: %s",
			req.Command, req.Username, req.Channel, req.Args)

		t, err := m.createTask(req, cmd)
		if err != nil {
			m.replyWithError(msg, fmt.Errorf("could not create job: %s", err))
			continue
		}

		m.wg.Add(1)
		m.tasksCh <- t
	}
}

func (m *Meeseeks) createTask(req request.Request, cmd command.Command) (task, error) {
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

func (m *Meeseeks) jobsLoop() {
	for t := range m.tasksCh {
		go func(t task) {
			job := t.job
			req := job.Request
			cmd := t.cmd

			m.replyWithHandshake(req, cmd)

			out, err := t.cmd.Execute(t.job)
			if err != nil {
				logrus.Errorf("Command '%s' from user '%s' failed execution with error: %s",
					req.Command, req.Username, err)
				m.replyWithCommandFailed(req, cmd, err, out)
				job.Finish(jobs.FailedStatus)
			} else {
				log.Infof("Command '%s' from user '%s' succeeded execution", req.Command,
					req.Username)
				m.replyWithSuccess(job.Request, cmd, out)
				job.Finish(jobs.SuccessStatus)
			}
			m.wg.Done()
		}(t)
	}
}
