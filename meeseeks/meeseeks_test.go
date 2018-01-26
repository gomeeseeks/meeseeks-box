package meeseeks_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/renstrom/dedent"

	"regexp"

	"github.com/pcarranza/meeseeks-box/config"
	"github.com/pcarranza/meeseeks-box/meeseeks"
	"github.com/pcarranza/meeseeks-box/meeseeks/message"
	"github.com/pcarranza/meeseeks-box/template"
	stubs "github.com/pcarranza/meeseeks-box/testingstubs"
)

type expectedMessage struct {
	TextMatcher string
	Channel     string
	IsIM        bool
}

func Test_BasicReplying(t *testing.T) {
	handshakeMatcher := fmt.Sprintf("^(%s)$", strings.Join(template.DefaultHandshakeMessages, "|"))

	tt := []struct {
		name     string
		user     string
		message  string
		channel  string
		expected []expectedMessage
	}{
		{
			name:    "basic case",
			user:    "myuser",
			message: "echo hello!",
			channel: "general",
			expected: []expectedMessage{
				expectedMessage{
					TextMatcher: handshakeMatcher,
					Channel:     "generalID",
					IsIM:        false,
				},
				expectedMessage{
					TextMatcher: "^<@myuser> .*\n```\nhello!\n```$",
					Channel:     "generalID",
					IsIM:        false,
				},
			},
		},
		{
			name:    "basic with cmds args",
			user:    "myuser",
			message: "args-echo hello!",
			channel: "general",
			expected: []expectedMessage{
				expectedMessage{
					TextMatcher: handshakeMatcher,
					Channel:     "generalID",
					IsIM:        false,
				},
				expectedMessage{
					TextMatcher: "^<@myuser> .*\n```\npre-message hello!\n```$",
					Channel:     "generalID",
					IsIM:        false,
				},
			},
		},
		{
			name:    "unknown command case",
			user:    "myuser",
			message: "unknown-command hello!",
			channel: "general",
			expected: []expectedMessage{
				expectedMessage{
					TextMatcher: "^<@myuser> Uuuh! no, I don't know how to do unknown-command$",
					Channel:     "generalID",
					IsIM:        false,
				},
			},
		},
		{
			name:    "no command to run",
			user:    "myuser",
			message: "",
			channel: "general",
			expected: []expectedMessage{
				expectedMessage{
					TextMatcher: "^<@myuser> Uuuh!, no, it failed :disappointed: No command to run$",
					Channel:     "generalID",
					IsIM:        false,
				},
			},
		},
		{
			name:    "disallowed command",
			user:    "myuser",
			message: "disallowed",
			channel: "general",
			expected: []expectedMessage{
				expectedMessage{
					TextMatcher: "<@myuser> Uuuuh, yeah! you are not allowed to do disallowed",
					Channel:     "generalID",
					IsIM:        false,
				},
			},
		},
		{
			name:    "fail command",
			user:    "myuser",
			message: "fail",
			channel: "general",
			expected: []expectedMessage{
				expectedMessage{
					TextMatcher: handshakeMatcher,
					Channel:     "generalID",
					IsIM:        false,
				},
				expectedMessage{
					TextMatcher: "^<@myuser> Uuuh!, no, it failed :disappointed: exit status 1$",
					Channel:     "generalID",
					IsIM:        false,
				},
			},
		},
	}

	client, cnf := stubs.NewHarness().
		WithConfig(dedent.Dedent(`
			---
			commands:
			  echo:
			    command: echo
			    auth_strategy: any
			    timeout: 5
			  fail:
			    command: false
			    auth_strategy: any
			  disallowed:
			    command: false
			    auth_strategy: none
			  args-echo:
			    command: echo
			    auth_strategy: any
			    timeout: 5
			    args: ["pre-message"]
			`)).Build()

	config.LoadConfig(cnf)

	m := meeseeks.New(client, cnf)
	messageCh := make(chan message.Message)
	go m.Start(messageCh)

	stubs.WithTmpDB(func() {
		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				t.Logf("starting test %s", tc.name)
				messageCh <- stubs.MessageStub{
					Text:      tc.message,
					Channel:   tc.channel,
					ChannelID: tc.channel + "ID",
					User:      tc.user,
				}
				t.Logf("message sent to channel on %s", tc.name)

				for _, expected := range tc.expected {
					t.Logf("reading replies from client on %s", tc.name)
					actual := <-client.Messages

					r, err := regexp.Compile(expected.TextMatcher)
					stubs.Must(t, "could not compile regex", err, expected.TextMatcher)

					if !r.MatchString(actual.Text) {
						t.Fatalf("Bad message, expected %s; got %s", expected.TextMatcher, actual.Text)
					}
					stubs.AssertEquals(t, expected.Channel, actual.Channel)
					stubs.AssertEquals(t, expected.IsIM, actual.IsIM)
				}
			})
		}
	})
	m.Shutdown()

}
