package request

import (
	"errors"

	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/message"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/request/parser"
)

// ErrNoCommandToRun is returned when a request can't identify a command to run
var ErrNoCommandToRun = errors.New("No command to run")

// Request is a structure that holds all the command execution request
type Request struct {
	Command  string
	Args     []string
	Username string
	ReplyTo  string
	Channel  string
	IsIM     bool
}

// FromMessage gets a message and generates a valid request from it
func FromMessage(msg message.Message) (Request, error) {
	args, err := parser.Parse(msg.GetText())
	if err != nil {
		return Request{}, err
	}

	if len(args) == 0 {
		return Request{}, ErrNoCommandToRun
	}

	return Request{
		Command:  args[0],
		Args:     args[1:],
		Username: msg.GetUsername(),
		ReplyTo:  msg.GetReplyTo(),
		Channel:  msg.GetChannel(),
		IsIM:     msg.IsIM(),
	}, nil
}
