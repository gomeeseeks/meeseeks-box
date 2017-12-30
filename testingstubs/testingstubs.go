package testingstubs

import (
	"bytes"
	"fmt"
)

// SentMessage is a message that has been sent through a client
type SentMessage struct {
	text    string
	channel string
	im      bool
}

func (s SentMessage) String() string {
	return fmt.Sprintf("channel: %s text: %s im: %t", s.channel, s.text, s.im)
}

// ClientStub is an extremely simple implementation of a client that only captures messages
// in an internal array
//
// It implements the Client interface
type ClientStub struct {
	messages []SentMessage
}

// NewClientStub returns a new empty but intialized Client stub
func NewClientStub() ClientStub {
	return ClientStub{
		messages: make([]SentMessage, 0),
	}
}

// Reply implements the meeseeks.Client.Reply interface
func (c *ClientStub) Reply(text, channel string) {
	c.messages = append(c.messages, SentMessage{text: text, channel: channel})
}

// ReplyIM implements the meeseeks.Client.ReplyIM interface
func (c *ClientStub) ReplyIM(text, user string) error {
	c.messages = append(c.messages, SentMessage{text: text, channel: user, im: true})
	return nil
}

func (c ClientStub) String() string {
	b := bytes.NewBufferString("")
	for _, m := range c.messages {
		b.WriteString(fmt.Sprintf("%s\n", m))
	}
	return b.String()
}

// Contains ensure that a given message has been issued through this container
func (c ClientStub) Contains(message string) bool {
	for _, m := range c.messages {
		if message == fmt.Sprint(m) {
			return true
		}
	}
	return false
}

// MessageStub is a simple stub that implements the Slack.Message interface
type MessageStub struct {
	Text    string
	Channel string
	User    string
}

// GetText implements the slack.Message.GetText interface
func (m MessageStub) GetText() string {
	return m.Text
}

// GetChannel implements the slack.Message.GetChannel interface
func (m MessageStub) GetChannel() string {
	return m.Channel
}

// GetUserFrom implements the slack.Message.GetUserFrom interface
func (m MessageStub) GetUserFrom() string {
	return fmt.Sprintf("<@%s>", m.User)
}
