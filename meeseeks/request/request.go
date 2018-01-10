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
	Command     string   `json:"Command"`
	Args        []string `json:"Arguments"`
	Username    string   `json:"Username"`
	UsernameID  string   `json:"UsernameID"`
	Channel     string   `json:"Channel"`
	ChannelID   string   `json:"CannelID"`
	ChannelLink string   `json:"CannelLink"`
	IsIM        bool     `json:"IsIM"`
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
		Command:     args[0],
		Args:        args[1:],
		Username:    msg.GetUsername(),
		UsernameID:  msg.GetUsernameID(),
		Channel:     msg.GetChannel(),
		ChannelID:   msg.GetChannelID(),
		ChannelLink: msg.GetChannelLink(),
		IsIM:        msg.IsIM(),
	}, nil
}
