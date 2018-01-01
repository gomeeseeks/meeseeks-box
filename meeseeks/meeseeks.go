package meeseeks

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"

	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	parser "gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/commandparser"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/template"
)

var (
	errCommandNotFound = errors.New("Could not find command")
	errNoCommandToRun  = errors.New("No command to run")
)

var builtinCommands = map[string]config.Command{
	"echo": config.Command{
		Cmd:          "echo",
		Timeout:      5,
		AuthStrategy: config.AuthStrategyAny,
	},
}

// Message interface to interact with an abstract message
type Message interface {
	GetText() string
	GetChannel() string
	GetUserFrom() string
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
		commands:  union(builtinCommands, conf.Commands),
		templates: template.DefaultTemplates(conf.Messages),
	}
}

// Process processes a received message
func (m Meeseeks) Process(message Message) {
	args, err := parser.ParseCommand(message.GetText())
	if err != nil {
		m.replyWithError(message, err, "can't parse command")
	}

	if len(args) == 0 {
		m.replyWithError(message, errNoCommandToRun, "")
		return
	}

	cmd, err := m.findCommand(args[0])
	if err != nil {
		m.replyWithUnknownCommand(message, args[0])
		return
	}

	m.replyWithHandshake(message)

	out, err := executeCommand(cmd, args[1:]...)
	if err != nil {
		m.replyWithError(message, err, out)
		return
	}

	m.replyWithSuccess(message, out)
}

func (m Meeseeks) replyWithHandshake(message Message) {
	msg, err := m.templates.RenderHandshake(message.GetUserFrom())
	if err != nil {
		log.Fatalf("could not render unknown command template %s", err)
	}

	m.client.Reply(msg, message.GetChannel())
}

func (m Meeseeks) replyWithUnknownCommand(message Message, cmd string) {

	msg, err := m.templates.RenderUnknownCommand(message.GetUserFrom(), cmd)
	if err != nil {
		log.Fatalf("could not render unknown command template %s", err)
	}

	m.client.Reply(msg, message.GetChannel())
}

func (m Meeseeks) replyWithError(message Message, err error, out string) {

	msg, err := m.templates.RenderFailure(message.GetUserFrom(), err.Error(), out)
	if err != nil {
		log.Fatalf("could not render failure template %s", err)
	}

	m.client.Reply(msg, message.GetChannel())
}

func (m Meeseeks) replyWithSuccess(message Message, out string) {
	msg, err := m.templates.RenderSuccess(message.GetUserFrom(), out)

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

func union(maps ...map[string]config.Command) map[string]config.Command {
	newMap := make(map[string]config.Command)
	for _, m := range maps {
		for k, v := range m {
			newMap[k] = v
		}
	}
	return newMap
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
