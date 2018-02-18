package meeseeks

import (
	"context"
	"errors"
	"time"
)

// Message interface to interact with an abstract message
type Message interface {
	// The text received without the matching portion
	GetText() string
	// The friendly name of the channel in which the message was issued
	GetChannel() string
	// The channel id used to build the channel link
	GetChannelID() string
	// The channel link is used in replies to show an hyperlink to the channel
	GetChannelLink() string
	// The friendly name of the user that has sent the message, used internally to match with groups and such
	GetUsername() string
	// The username id of the user that has sent the message, used in replies so they notify the user
	GetUserID() string
	// The user link returns a link to the user
	GetUserLink() string
	// IsIM
	IsIM() bool
}

// Request is a structure that holds a command execution request
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

// Job represents a single job
type Job struct {
	ID        uint64    `json:"ID"`
	Request   Request   `json:"Request"`
	StartTime time.Time `json:"StartTime"`
	EndTime   time.Time `json:"EndTime"`
	Status    string    `json:"Status"`
}

// APIToken is a persisted API token
type APIToken struct {
	TokenID     string    `json:"token"`
	UserLink    string    `json:"userLink"`
	ChannelLink string    `json:"channelLink"`
	Text        string    `json:"text"`
	CreatedOn   time.Time `json:"created_on"`
}

// JobLog represents all the logging information of a given Job
type JobLog struct {
	Error  string
	Output string
}

// GetError returns nil or an error depending on the current JobLog setup
func (j JobLog) GetError() error {
	if j.Error == "" {
		return nil
	}
	return errors.New(j.Error)
}

// Defaults for commands
const (
	DefaultCommandTimeout = 60 * time.Second
)

// Command is the base interface for any command
type Command interface {
	Execute(context.Context, Job) (string, error)
	Cmd() string
	HasHandshake() bool
	Templates() map[string]string
	AuthStrategy() string
	AllowedGroups() []string
	Args() []string
	Timeout() time.Duration
	Help() Help
	Record() bool
}

// Help is the base interface for any command help
type Help interface {
	GetSummary() string
	GetArgs() []string
}
