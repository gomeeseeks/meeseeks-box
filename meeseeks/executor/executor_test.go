package executor_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/gomeeseeks/meeseeks-box/meeseeks/executor"
	"github.com/gomeeseeks/meeseeks-box/messenger"
	"github.com/gomeeseeks/meeseeks-box/mocks"
	"github.com/gomeeseeks/meeseeks-box/template"
	"github.com/renstrom/dedent"
	"github.com/sirupsen/logrus"
)

type expectedMessage struct {
	TextMatcher string
	Channel     string
	IsIM        bool
}

func Test_MeeseeksInteractions(t *testing.T) {
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
				{
					TextMatcher: handshakeMatcher,
					Channel:     "generalID",
					IsIM:        false,
				},
				{
					TextMatcher: "^<@myuser> .*\n```\nhello!```$",
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
				{
					TextMatcher: handshakeMatcher,
					Channel:     "generalID",
					IsIM:        false,
				},
				{
					TextMatcher: "^<@myuser> .*\n```\npre-message hello!```$",
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
				{
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
				{
					TextMatcher: "^<@myuser> Uuuh!, no, it failed :disappointed: no command to run$",
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
				{
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
				{
					TextMatcher: handshakeMatcher,
					Channel:     "generalID",
					IsIM:        false,
				},
				{
					TextMatcher: "^<@myuser> Uuuh!, no, it failed :disappointed: exit status 1$",
					Channel:     "generalID",
					IsIM:        false,
				},
			},
		},
	}

	mocks.WithTmpDB(func(dbpath string) {
		client := mocks.NewHarness().
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
			`)).WithDBPath(dbpath).Load()

		msgs, err := messenger.Listen(client)
		if err != nil {
			t.Fatalf("could not create listener: %s", err)
		}
		m := executor.New(client, msgs)
		go m.Start()

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				logrus.Infof("starting test %s", tc.name)
				client.MessagesCh() <- mocks.MessageStub{
					Text:    tc.message,
					Channel: tc.channel,
					User:    tc.user,
				}
				logrus.Infof("message sent to channel on %s", tc.name)

				logrus.Infof("reading replies from client on %s", tc.name)
				for _, expected := range tc.expected {
					logrus.Infof("expecting message %#v", expected)

					actual := <-client.MessagesSent

					logrus.Infof("got message %#v", actual)

					r, err := regexp.Compile(expected.TextMatcher)
					mocks.Must(t, "could not compile regex", err, expected.TextMatcher)

					if !r.MatchString(actual.Text) {
						t.Fatalf("Bad message, expected %s; got %s", expected.TextMatcher, actual.Text)
					}
					mocks.AssertEquals(t, expected.Channel, actual.Channel)
					mocks.AssertEquals(t, expected.IsIM, actual.IsIM)
				}
			})
		}
		m.Shutdown()
	})

}
