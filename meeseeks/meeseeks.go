package meeseeks

import (
	log "github.com/sirupsen/logrus"

	"gitlab.com/mr-meeseeks/meeseeks-box/auth"
	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/command"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/message"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/template"
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
	commands  command.Commands
	templates *template.TemplatesBuilder
}

// New creates a new Meeseeks service
func New(client Client, conf config.Config) Meeseeks {
	cmds, _ := command.New(conf) // TODO handle the error
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
	req, err := command.RequestFromMessage(msg)
	if err != nil {
		log.Debugf("Failed to parse message '%s' as a command: %s", msg.GetText(), err)
		m.replyWithError(msg, err, "")
		return
	}

	cmd, err := m.commands.Find(req.Command)
	if err == command.ErrCommandNotFound {
		m.replyWithUnknownCommand(msg, req.Command)
		return
	}
	if err = auth.Check(req.Command, cmd.ConfiguredCommand()); err != nil {
		m.replyWithUnauthorizedCommand(cmd, req.Command, msg)
		return
	}

	log.Infof("Accepted command '%s' from user '%s' with args: %s", req.Command, req.Username, req.Args)
	m.replyWithHandshake(cmd, msg)

	out, err := cmd.Execute(req.Args...)
	if err != nil {
		log.Errorf("Command '%s' from user '%s' failed execution with error: %s",
			cmd, req.Username, err)
		m.replyWithCommandFailed(cmd, msg, err, out)
		return
	}

	m.replyWithSuccess(cmd, msg, out)
	log.Infof("Command '%s' from user '%s' succeeded execution", cmd, req.Username)
}

func (m Meeseeks) buildTemplatesFor(cmd command.Command) template.Templates {
	return m.templates.Clone().WithTemplates(cmd.ConfiguredCommand().Templates).Build()
}

func (m Meeseeks) replyWithError(message message.Message, err error, out string) {
	msg, err := m.templates.Build().RenderFailure(message.GetReplyTo(), err.Error(), out)
	if err != nil {
		log.Fatalf("could not render failure template: %s", err)
	}

	m.client.Reply(msg, m.config.Colors.Error, message.GetChannel())
}

func (m Meeseeks) replyWithUnknownCommand(message message.Message, cmd string) {
	log.Debugf("Could not find command '%s' in the command registry", cmd)

	msg, err := m.templates.Build().RenderUnknownCommand(message.GetReplyTo(), cmd)
	if err != nil {
		log.Fatalf("could not render unknown command template: %s", err)
	}

	m.client.Reply(msg, m.config.Colors.Error, message.GetChannel())
}

func (m Meeseeks) replyWithHandshake(cmd command.Command, message message.Message) {
	if !cmd.HasHandshake() {
		return
	}
	msg, err := m.buildTemplatesFor(cmd).RenderHandshake(message.GetReplyTo())
	if err != nil {
		log.Fatalf("could not render unknown command template: %s", err)
	}

	m.client.Reply(msg, m.config.Colors.Info, message.GetChannel())
}

func (m Meeseeks) replyWithUnauthorizedCommand(cmd command.Command, commandName string, message message.Message) {
	log.Debugf("User %s is not allowed to run command '%s'", message.GetUsername(), cmd)

	msg, err := m.buildTemplatesFor(cmd).RenderUnauthorizedCommand(message.GetReplyTo(), commandName)
	if err != nil {
		log.Fatalf("could not render unathorized command template %s", err)
	}

	m.client.Reply(msg, m.config.Colors.Error, message.GetChannel())
}

func (m Meeseeks) replyWithCommandFailed(cmd command.Command, message message.Message, err error, out string) {
	msg, err := m.buildTemplatesFor(cmd).RenderFailure(message.GetReplyTo(), err.Error(), out)
	if err != nil {
		log.Fatalf("could not render failure template %s", err)
	}

	m.client.Reply(msg, m.config.Colors.Error, message.GetChannel())
}

func (m Meeseeks) replyWithSuccess(cmd command.Command, message message.Message, out string) {
	msg, err := m.buildTemplatesFor(cmd).RenderSuccess(message.GetReplyTo(), out)

	if err != nil {
		log.Fatalf("could not render success template %s", err)
	}

	m.client.Reply(msg, m.config.Colors.Success, message.GetChannel())
}
