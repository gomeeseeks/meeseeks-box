package meeseeks

import "fmt"

// Message interface to interact with an abstract message
type Message interface {
	GetText() string
	GetChannel() string
	GetUserFrom() string
	GetUserFromID() string
}

// Client interface that provides a way of replying to messages on a channel
type Client interface {
	Reply(text, channel string)
	ReplyIM(text, user string) error
}

// ProcessMessage processes a received message
func ProcessMessage(m Message, c Client) {
	c.Reply(fmt.Sprintf("%s echo: %s", m.GetUserFrom(), m.GetText()), m.GetChannel())
}
