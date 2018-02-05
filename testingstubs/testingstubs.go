package testingstubs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/pcarranza/meeseeks-box/config"
	"github.com/pcarranza/meeseeks-box/db"
	"github.com/pcarranza/meeseeks-box/meeseeks/message"

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
	cnf    string
	dbpath string
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

// WithEchoCommand returns a harness configured to have an echo command available
func (h Harness) WithEchoCommand() Harness {
	h.cnf = `---
commands:
  echo:
    command: echo
    auth_strategy: any
    timeout: 5
`
	return h
}

// WithConfig allows to change the configuration string
func (h Harness) WithConfig(c string) Harness {
	h.cnf = c
	return h
}

// WithDBPath provides a dabatase filepath for the testing harness
func (h Harness) WithDBPath(dbpath string) Harness {
	h.dbpath = dbpath
	return h
}

// Build creates a clientStub and a configuration based on the provided one
func (h Harness) Load() (ClientStub, config.Config) {
	c, err := config.New(strings.NewReader(h.cnf))
	if err != nil {
		log.Fatalf("Could not build test harness: %s", err)
	}
	if h.dbpath != "" {
		c.Database = db.DatabaseConfig{
			Path:    h.dbpath,
			Mode:    0600,
			Timeout: 2 * time.Second,
		}
	}
	if err := config.LoadConfig(c); err != nil {
		fmt.Printf("Failed to load configuration: %s", err)
		return ClientStub{}, c
	}
	return newClientStub(), c
}

// ClientStub is an extremely simple implementation of a client that only captures messages
// in an internal array
//
// It implements the Client interface
type ClientStub struct {
	MessagesSent     chan SentMessage
	receivedMessages chan message.Message
}

// NewClientStub returns a new empty but intialized Client stub
func newClientStub() ClientStub {
	return ClientStub{
		MessagesSent:     make(chan SentMessage),
		receivedMessages: make(chan message.Message),
	}
}

// Reply implements the meeseeks.Client.Reply interface
func (c ClientStub) Reply(text, color, channel string) error {
	c.MessagesSent <- SentMessage{Text: text, Color: color, Channel: channel}
	return nil
}

// ReplyIM implements the meeseeks.Client.ReplyIM interface
func (c ClientStub) ReplyIM(text, color, user string) error {
	c.MessagesSent <- SentMessage{Text: text, Color: color, Channel: user, IsIM: true}
	return nil
}

// MessagesCh implements meeseeks.Client.MessagesCh interface
func (c ClientStub) MessagesCh() chan message.Message {
	return c.receivedMessages
}

// ListenMessages listens for messages and then passes it to the sent channel
func (c ClientStub) ListenMessages(ch chan<- message.Message) {
	for m := range c.receivedMessages {
		ch <- m
	}
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

// GetUserLink implements the slack.Message.GetUserFrom interface
func (m MessageStub) GetUserLink() string {
	return fmt.Sprintf("<@%s>", m.User)
}

// GetUsername implements the slack.Message.GetUsername interface
func (m MessageStub) GetUsername() string {
	return m.User
}

// GetUserID implements the slack.Message.GetUsername interface
func (m MessageStub) GetUserID() string {
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

// AssertMatches Helper function for asserting that a value is what we expect
func AssertMatches(t *testing.T, expected, actual string) {
	r, err := regexp.Compile(expected)
	Must(t, fmt.Sprintf("regex %s does not compile", expected), err)
	if !r.Match([]byte(actual)) {
		t.Fatalf("Value does not match expected,\nexpected %#v;\ngot %#v", expected, actual)
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
func WithTmpDB(f func(dbpath string)) error {
	tmpdir, err := ioutil.TempDir("", "meeseeks")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)

	dbpath := path.Join(tmpdir, "meeseeks.db")
	db.Configure(db.DatabaseConfig{
		Path:    dbpath,
		Mode:    0600,
		Timeout: 1 * time.Second,
	})

	f(dbpath)

	return nil
}

type MetadataStub struct {
	IM bool
}

func (m MetadataStub) GetChannelLink(channelID string) string {
	return fmt.Sprintf("<#%s>", channelID)
}

func (m MetadataStub) GetChannelID(channelID string) string {
	return channelID
}

func (m MetadataStub) GetChannel(channelID string) string {
	return fmt.Sprintf("name: %s", channelID)
}

func (m MetadataStub) GetUserLink(userID string) string {
	return fmt.Sprintf("<@%s>", userID)
}

func (m MetadataStub) GetUserID(userID string) string {
	return userID
}

func (m MetadataStub) GetUsername(userID string) string {
	return fmt.Sprintf("name: %s", userID)
}

func (m MetadataStub) IsIM(_ string) bool {
	return m.IM
}
