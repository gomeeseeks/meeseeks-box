package meeseeks

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/auth"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/commandparser"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/commands"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/template"
)

var (
	errCommandNotFound = errors.New("Could not find command")
	errNoCommandToRun  = errors.New("No command to run")
)

// Message interface to interact with an abstract message
type Message interface {
	GetText() string
	GetChannel() string
	GetReplyTo() string
	GetUsername() string
	IsIM() bool
}

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
	templates template.Templates
}

// New creates a new Meeseeks service
func New(client Client, conf config.Config) Meeseeks {
	cmds, _ := commands.New(conf) // TODO handle the error
	return Meeseeks{
		client:    client,
		config:    conf,
		commands:  cmds,
		templates: template.DefaultTemplates(conf.Messages),
	}
}

// Process processes a received message
func (m Meeseeks) Process(message Message) {
	args, err := commandparser.ParseCommand(message.GetText())
	if err != nil {
		log.Debugf("Failed to parse message '%s' as a command: %s", message.GetText(), err)
		m.replyWithError(message, err, "can't parse command")
	}

	if len(args) == 0 {
		log.Debugf("Could not find any command in message '%s'", message.GetText())
		m.replyWithError(message, errNoCommandToRun, "")
		return
	}

	cmd := args[0]
	command, err := m.commands.Find(cmd)
	if err == commands.ErrCommandNotFound {
		m.replyWithUnknownCommand(message, args[0])
		return
	}
	if err = auth.Check(message.GetUsername(), command.ConfiguredCommand()); err != nil {
		m.replyWithUnauthorizedCommand(message, args[0])
		return
	}

	log.Infof("Accepted command '%s' from user %s with args: %s", cmd, message.GetUsername(), args[1:])
	if command.HasHandshake() {
		m.replyWithHandshake(message)
	}

	out, err := command.Execute(args[1:]...)
	if err != nil {
		log.Errorf("Command '%s' from user %s failed execution with error: %s",
			cmd, message.GetUsername(), err)
		m.replyWithError(message, err, out)
		return
	}

	m.replyWithSuccess(message, out)
	log.Infof("Command '%s' from user %s succeeded execution", cmd, message.GetUsername())
}

func (m Meeseeks) replyWithHandshake(message Message) {
	msg, err := m.templates.RenderHandshake(message.GetReplyTo())
	if err != nil {
		log.Fatalf("could not render unknown command template %s", err)
	}

	m.client.Reply(msg, m.config.Colors.Info, message.GetChannel())
}

func (m Meeseeks) replyWithUnknownCommand(message Message, cmd string) {
	log.Debugf("Could not find command '%s' in the registered commands", cmd)

	msg, err := m.templates.RenderUnknownCommand(message.GetReplyTo(), cmd)
	if err != nil {
		log.Fatalf("could not render unknown command template %s", err)
	}

	m.client.Reply(msg, m.config.Colors.Error, message.GetChannel())
}

func (m Meeseeks) replyWithUnauthorizedCommand(message Message, cmd string) {
	log.Debugf("User %s is not allowed to run command '%s'", message.GetUsername(), cmd)

	msg, err := m.templates.RenderUnauthorizedCommand(message.GetReplyTo(), cmd)
	if err != nil {
		log.Fatalf("could not render unathorized command template %s", err)
	}

	m.client.Reply(msg, m.config.Colors.Error, message.GetChannel())
}

func (m Meeseeks) replyWithError(message Message, err error, out string) {

	msg, err := m.templates.RenderFailure(message.GetReplyTo(), err.Error(), out)
	if err != nil {
		log.Fatalf("could not render failure template %s", err)
	}

	m.client.Reply(msg, m.config.Colors.Error, message.GetChannel())
}

func (m Meeseeks) replyWithSuccess(message Message, out string) {
	msg, err := m.templates.RenderSuccess(message.GetReplyTo(), out)

	if err != nil {
		log.Fatalf("could not render success template %s", err)
	}

	m.client.Reply(msg, m.config.Colors.Success, message.GetChannel())
}
