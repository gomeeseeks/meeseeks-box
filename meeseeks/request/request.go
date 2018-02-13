package request

import (
	"errors"

	"github.com/gomeeseeks/meeseeks-box/meeseeks/message"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/request/parser"
	"github.com/sirupsen/logrus"
)

// ErrNoCommandToRun is returned when a request can't identify a command to run
var ErrNoCommandToRun = errors.New("No command to run")

// Request is a structure that holds all the command execution request
type Request struct {
	Command     string   `json:"Command"`
	Args        []string `json:"Arguments"`
	Username    string   `json:"Username"`
	UserID      string   `json:"UserID"`
	UserLink    string   `json:"UserLink"`
	Channel     string   `json:"Channel"`
	ChannelID   string   `json:"CannelID"`
	ChannelLink string   `json:"CannelLink"`
	IsIM        bool     `json:"IsIM"`
}

// FromMessage gets a message and generates a valid request from it
func FromMessage(msg message.Message) (Request, error) {
	args, err := parser.Parse(msg.GetText())
	logrus.Debugf("Command %s parsed as %s", msg.GetText(), args)

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
		UserID:      msg.GetUserID(),
		UserLink:    msg.GetUserLink(),
		Channel:     msg.GetChannel(),
		ChannelID:   msg.GetChannelID(),
		ChannelLink: msg.GetChannelLink(),
		IsIM:        msg.IsIM(),
	}, nil
}
