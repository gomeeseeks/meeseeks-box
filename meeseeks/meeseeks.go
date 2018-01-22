package meeseeks

import (
	"github.com/pcarranza/meeseeks-box/jobs"
	log "github.com/sirupsen/logrus"

	"github.com/pcarranza/meeseeks-box/auth"
	"github.com/pcarranza/meeseeks-box/config"
	"github.com/pcarranza/meeseeks-box/meeseeks/commands"
	"github.com/pcarranza/meeseeks-box/meeseeks/message"
	"github.com/pcarranza/meeseeks-box/meeseeks/request"
	"github.com/pcarranza/meeseeks-box/meeseeks/template"
)

// Client interface that provides a way of replying to messages on a channel
type Client interface {
	Reply(text, color, channel string) error
	ReplyIM(text, color, user string) error
}

// Meeseeks is the command execution engine
type Meeseeks struct {
	client    Client
	config    config.Config
	commands  commands.Commands
	templates *template.TemplatesBuilder
}

// New creates a new Meeseeks service
func New(client Client, conf config.Config) Meeseeks {
	cmds, _ := commands.New(conf) // TODO handle the error
	templatesBuilder := template.NewBuilder().WithMessages(conf.Messages)
	return Meeseeks{
		client:    client,
		config:    conf,
		commands:  cmds,
		templates: templatesBuilder,
	}
}

// Process processes a received message
func (m Meeseeks) Process(msg message.Message) {
	req, err := request.FromMessage(msg)
	if err != nil {
		log.Debugf("Failed to parse message '%s' as a command: %s", msg.GetText(), err)
		m.replyWithInvalidMessage(msg, err)
		return
	}

	cmd, err := m.commands.Find(req.Command)
	if err == commands.ErrCommandNotFound {
		m.replyWithUnknownCommand(req)
		return
	}
	if err = auth.Check(req.Username, cmd); err != nil {
		m.replyWithUnauthorizedCommand(req, cmd)
		return
	}

	log.Infof("Accepted command '%s' from user '%s' on channel '%s' with args: %s",
		req.Command, req.Username, req.Channel, req.Args)

	var j jobs.Job
	if cmd.Record() {
		j, err = jobs.Create(req)
		if err != nil {
			log.Errorf("could not create job: %s", err)
		}
	} else {
		j = jobs.NullJob(req)
	}

	m.replyWithHandshake(req, cmd)

	out, err := cmd.Execute(j)
	if err != nil {
		log.Errorf("Command '%s' from user '%s' failed execution with error: %s",
			req.Command, req.Username, err)
		m.replyWithCommandFailed(req, cmd, err, out)
		j.Finish(jobs.FailedStatus)
		return
	}

	log.Infof("Command '%s' from user '%s' succeeded execution", req.Command,
		req.Username)
	m.replyWithSuccess(j.Request, cmd, out)
	j.Finish(jobs.SuccessStatus)
}
