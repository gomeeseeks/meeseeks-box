package testingstubs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pcarranza/meeseeks-box/config"
	"github.com/pcarranza/meeseeks-box/db"

	log "github.com/sirupsen/logrus"
)

// SentMessage is a message that has been sent through a client
type SentMessage struct {
	Text    string
	Channel string
	Color   string
	IsIM    bool
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
func (c ClientStub) Reply(text, color, channel string) error {
	c.Messages <- SentMessage{Text: text, Color: color, Channel: channel}
	return nil
}

// ReplyIM implements the meeseeks.Client.ReplyIM interface
func (c ClientStub) ReplyIM(text, color, user string) error {
	c.Messages <- SentMessage{Text: text, Color: color, Channel: user, IsIM: true}
	return nil
}

// MessageStub is a simple stub that implements the Slack.Message interface
type MessageStub struct {
	Text      string
	Channel   string
	User      string
	UserID    string
	ChannelID string
	IM        bool
}

// GetText implements the slack.Message.GetText interface
func (m MessageStub) GetText() string {
	return m.Text
}

// GetChannel implements the slack.Message.GetChannel interface
func (m MessageStub) GetChannel() string {
	return m.Channel
}

// GetChannelID implements the slack.Message.GetUserFrom interface
func (m MessageStub) GetChannelID() string {
	return m.Channel + "ID"
}

// GetChannelLink implements the slack.Message.GetUserFrom interface
func (m MessageStub) GetChannelLink() string {
	return m.Channel + "Link"
}

// GetUsernameID implements the slack.Message.GetUserFrom interface
func (m MessageStub) GetUsernameID() string {
	return fmt.Sprintf("<@%s>", m.User)
}

// GetUsername implements the slack.Message.GetUsername interface
func (m MessageStub) GetUsername() string {
	return m.User
}

// IsIM implements the slack.Message.IsIM
func (m MessageStub) IsIM() bool {
	return m.IM
}

// AssertEquals Helper function for asserting that a value is what we expect
func AssertEquals(t *testing.T, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Value is not as expected,\nexpected %#v;\ngot %#v", expected, actual)
	}
}

// Must is a helper function that allows to fail the test with a message if there's an error
func Must(t *testing.T, message string, err error, additionalDetails ...string) {
	if err != nil {
		m := []string{fmt.Sprintf("%s %s", message, err)}
		m = append(m, additionalDetails...)
		t.Fatal(m)
	}
}

// WithTmpDB creates a temporary database in which to run persistence tests
func WithTmpDB(f func()) error {
	tmpdir, err := ioutil.TempDir("", "meeseeks")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)

	db.Configure(config.Config{
		Database: config.Database{
			Path:    path.Join(tmpdir, "meeseeks.db"),
			Mode:    0600,
			Timeout: time.Second * 1,
		},
	})
	f()
	return nil
}
