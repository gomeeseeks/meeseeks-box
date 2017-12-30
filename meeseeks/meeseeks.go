package meeseeks

import (
	"fmt"

	"gitlab.com/mr-meeseeks/meeseeks-box/config"
)

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
	client Client
	config config.Config
}

// New creates a new Meeseeks service
func New(client Client, config config.Config) Meeseeks {
	return Meeseeks{
		client: client,
		config: config,
	}
}

// Process processes a received message
func (m Meeseeks) Process(message Message) {
	m.client.Reply(fmt.Sprintf("%s echo: %s", message.GetUserFrom(), message.GetText()), message.GetChannel())
}
