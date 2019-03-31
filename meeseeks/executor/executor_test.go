package executor_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"gitlab.com/yakshaving.art/meeseeks-box/meeseeks"
	"gitlab.com/yakshaving.art/meeseeks-box/meeseeks/executor"
	"gitlab.com/yakshaving.art/meeseeks-box/mocks"
	"gitlab.com/yakshaving.art/meeseeks-box/text/template"
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
		name      string
		userLink  string
		cmd       string
		args      []string
		channelID string
		expected  []expectedMessage
	}{
		{
			name:      "basic case",
			userLink:  "<@myuser>",
			channelID: "generalID",
			cmd:       "echo",
			args:      []string{"hello!"},
			expected: []expectedMessage{
				{
					TextMatcher: handshakeMatcher,
					Channel:     "generalID",
					IsIM:        false,
				},
				{
					TextMatcher: "^<@myuser> .*\n```\nhello!\n```$",
					Channel:     "generalID",
					IsIM:        false,
				},
			},
		},
		{
			name:      "help command",
			userLink:  "<@myuser>",
			channelID: "generalID",
			cmd:       "help",
			args:      []string{},
			expected: []expectedMessage{
				{
					TextMatcher: "^<@myuser> .*\n```\n- args-echo: \n- disallowed: \n- echo: \n- fail: \n```$",
					Channel:     "generalID",
					IsIM:        false,
				},
			},
		},
		{
			name:      "basic with cmds args",
			userLink:  "<@myuser>",
			channelID: "generalID",
			cmd:       "args-echo",
			args:      []string{"hello!"},
			expected: []expectedMessage{
				{
					TextMatcher: handshakeMatcher,
					Channel:     "generalID",
					IsIM:        false,
				},
				{
					TextMatcher: "^<@myuser> .*\n```\npre-message hello!\n```$",
					Channel:     "generalID",
					IsIM:        false,
				},
			},
		},
		{
			name:      "unknown command case",
			userLink:  "<@myuser>",
			channelID: "generalID",
			cmd:       "unknown-command",
			args:      []string{"hello!"},
			expected: []expectedMessage{
				{
					TextMatcher: "^<@myuser> Uuuh! no, I don't know how to do unknown-command$",
					Channel:     "generalID",
					IsIM:        false,
				},
			},
		},
		{
			name:      "disallowed command",
			userLink:  "<@myuser>",
			channelID: "generalID",
			cmd:       "disallowed",
			args:      []string{},
			expected: []expectedMessage{
				{
					TextMatcher: "<@myuser> Uuuuh, yeah! you are not allowed to do disallowed",
					Channel:     "generalID",
					IsIM:        false,
				},
			},
		},
		{
			name:      "fail command",
			userLink:  "<@myuser>",
			channelID: "generalID",
			cmd:       "fail",
			args:      []string{},
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

		e := executor.New(executor.Args{
			ChatClient:          client,
			WithBuiltinCommands: true,
			ConcurrentTaskCount: 1,
		})
		e.ListenTo(client)

		go e.Run()

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				logrus.Infof("starting test %s", tc.name)
				client.RequestsCh <- meeseeks.Request{
					Command:   tc.cmd,
					Args:      tc.args,
					UserLink:  tc.userLink,
					ChannelID: tc.channelID,
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
		e.Shutdown()
	})

}
