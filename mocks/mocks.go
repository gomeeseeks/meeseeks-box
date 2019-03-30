package mocks

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

	"gitlab.com/yakshaving.art/meeseeks-box/config"
	"gitlab.com/yakshaving.art/meeseeks-box/meeseeks"
	"gitlab.com/yakshaving.art/meeseeks-box/persistence/db"
	"gitlab.com/yakshaving.art/meeseeks-box/text/formatter"

	"github.com/sirupsen/logrus"
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
		logrus.Fatalf("Failed to read configuration file %s: %s", f, err)
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

// Load creates a clientStub and a configuration based on the provided one
func (h Harness) Load() ClientStub {
	c, err := config.New(strings.NewReader(h.cnf))
	if err != nil {
		logrus.Fatalf("Could not build test harness: %s", err)
	}
	if h.dbpath != "" {
		c.Database = db.DatabaseConfig{
			Path:    h.dbpath,
			Mode:    0600,
			Timeout: 2 * time.Second,
		}
	}
	if err := config.LoadConfiguration(c); err != nil {
		fmt.Printf("Failed to load configuration: %s", err)
		return ClientStub{}
	}
	return newClientStub()
}

// ClientStub is an extremely simple implementation of a client that only captures messages
// in an internal array
//
// It implements the Client interface
type ClientStub struct {
	MessagesSent chan SentMessage
	RequestsCh   chan meeseeks.Request
}

// NewClientStub returns a new empty but intialized Client stub
func newClientStub() ClientStub {
	return ClientStub{
		MessagesSent: make(chan SentMessage),
		RequestsCh:   make(chan meeseeks.Request),
	}
}

// Reply implements the meeseeks.Client.Reply interface
func (c ClientStub) Reply(r formatter.Reply) {
	logrus.Infof("sending reply %#v to client", r)
	text, err := r.Render()
	if err != nil {
		logrus.Error(err)
	}
	c.MessagesSent <- SentMessage{Text: text, Channel: r.ChannelID()}
}

// Listen listens for requests and then passes them to the passed in channel
func (c ClientStub) Listen(ch chan<- meeseeks.Request) {
	for m := range c.RequestsCh {
		ch <- m
	}
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

// EnricherStub provides a stub object that implements the Metadata interface
type EnricherStub struct {
	IM bool
}

// ParseChannelLink implements the Metadata interface
func (m EnricherStub) ParseChannelLink(channelLink string) (string, error) {
	return strings.Replace(channelLink, "Link", "", -1), nil
}

// ParseUserLink implements the Metadata interface
func (m EnricherStub) ParseUserLink(userLink string) (string, error) {
	return strings.Replace(userLink, "Link", "", -1), nil
}

// GetChannelLink implements the Metadata interface
func (m EnricherStub) GetChannelLink(channelID string) string {
	return fmt.Sprintf("<#%s>", channelID)
}

// GetChannelID implements the Metadata interface
func (m EnricherStub) GetChannelID(channelID string) string {
	return channelID
}

// GetChannel implements the Metadata interface
func (m EnricherStub) GetChannel(channelID string) string {
	return fmt.Sprintf("name: %s", channelID)
}

// GetUserLink implements the Metadata interface
func (m EnricherStub) GetUserLink(userID string) string {
	return fmt.Sprintf("<@%s>", userID)
}

// GetUserID implements the Metadata interface
func (m EnricherStub) GetUserID(userID string) string {
	return userID
}

// GetUsername implements the Metadata interface
func (m EnricherStub) GetUsername(userID string) string {
	return fmt.Sprintf("name: %s", userID)
}

// IsIM implements the Metadata interface
func (m EnricherStub) IsIM(_ string) bool {
	return m.IM
}
