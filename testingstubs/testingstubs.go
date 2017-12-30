package testingstubs

import (
	"fmt"
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
	"gitlab.com/mr-meeseeks/meeseeks-box/config"
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

// Harness is a builder that helps out testing meeseeks
type Harness struct {
	cnf string
}

// NewHarness returns a new empty harness
func NewHarness() Harness {
	return Harness{}
}

// WithConfigFile allows to read a configuration file
func (h Harness) WithConfigFile(f string) Harness {
	s, err := ioutil.ReadFile(f)
	if err != nil {
		log.Fatalf("Failed to read configuration file %s: %s", f, err)
	}
	h.cnf = string(s)
	return h
}

// WithConfig allows to change the configuration string
func (h Harness) WithConfig(c string) Harness {
	h.cnf = c
	return h
}

// Build creates a clientStub and a configuration based on the provided one
func (h Harness) Build() (ClientStub, config.Config) {
	c, err := config.New(strings.NewReader(h.cnf))
	if err != nil {
		log.Fatalf("Could not build test harness: %s", err)
	}
	return newClientStub(), c
}

// ClientStub is an extremely simple implementation of a client that only captures messages
// in an internal array
//
// It implements the Client interface
type ClientStub struct {
	Messages chan SentMessage
}

// NewClientStub returns a new empty but intialized Client stub
func newClientStub() ClientStub {
	return ClientStub{
		Messages: make(chan SentMessage),
	}
}

// Reply implements the meeseeks.Client.Reply interface
func (c ClientStub) Reply(text, channel string) {
	c.Messages <- SentMessage{text: text, channel: channel}
}

// ReplyIM implements the meeseeks.Client.ReplyIM interface
func (c ClientStub) ReplyIM(text, user string) error {
	c.Messages <- SentMessage{text: text, channel: user, im: true}
	return nil
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
