package meeseeks

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"

	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/auth"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/commandparser"
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
}

// Client interface that provides a way of replying to messages on a channel
type Client interface {
	Reply(text, channel string)
	ReplyIM(text, user string) error
}

// Meeseeks is the command execution engine
type Meeseeks struct {
	client    Client
	config    config.Config
	commands  map[string]config.Command
	templates template.Templates
}

// New creates a new Meeseeks service
func New(client Client, conf config.Config) Meeseeks {
	return Meeseeks{
		client:    client,
		config:    conf,
		commands:  conf.GetCommands(),
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
	command, err := m.findCommand(cmd)
	if err != nil {
		m.replyWithUnknownCommand(message, cmd)
		return
	}
	if !auth.IsAllowed(message.GetUsername(), command) {
		m.replyWithUnauthorizedCommand(message, command.Cmd)
		return
	}

	m.replyWithHandshake(message)
	log.Infof("Accepted command '%s' from user %s with args: %s", cmd, message.GetUsername(), args[1:])

	out, err := executeCommand(command, args[1:]...)
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

	m.client.Reply(msg, message.GetChannel())
}

func (m Meeseeks) replyWithUnknownCommand(message Message, cmd string) {
	log.Debugf("Could not find command '%s' in the registered commands", cmd)

	msg, err := m.templates.RenderUnknownCommand(message.GetReplyTo(), cmd)
	if err != nil {
		log.Fatalf("could not render unknown command template %s", err)
	}

	m.client.Reply(msg, message.GetChannel())
}

func (m Meeseeks) replyWithUnauthorizedCommand(message Message, cmd string) {
	log.Debugf("User %s is not allowed to run command '%s'", message.GetUsername(), cmd)

	msg, err := m.templates.RenderUnauthorizedCommand(message.GetReplyTo(), cmd)
	if err != nil {
		log.Fatalf("could not render unathorized command template %s", err)
	}

	m.client.Reply(msg, message.GetChannel())
}

func (m Meeseeks) replyWithError(message Message, err error, out string) {

	msg, err := m.templates.RenderFailure(message.GetReplyTo(), err.Error(), out)
	if err != nil {
		log.Fatalf("could not render failure template %s", err)
	}

	m.client.Reply(msg, message.GetChannel())
}

func (m Meeseeks) replyWithSuccess(message Message, out string) {
	msg, err := m.templates.RenderSuccess(message.GetReplyTo(), out)

	if err != nil {
		log.Fatalf("could not render success template %s", err)
	}

	m.client.Reply(msg, message.GetChannel())
}

func (m Meeseeks) findCommand(command string) (config.Command, error) {
	cmd, ok := m.commands[command]
	if !ok {
		return config.Command{}, fmt.Errorf("%s '%s'", errCommandNotFound, command)
	}
	return cmd, nil
}

func executeCommand(cmd config.Command, args ...string) (string, error) {
	timeout := time.Duration(cmd.Timeout) * time.Second
	args = append(cmd.Args, args...)

	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()

	c := exec.CommandContext(ctx, cmd.Cmd, args...)
	out, err := c.CombinedOutput()

	return string(out), err
}
