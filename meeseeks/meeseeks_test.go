package meeseeks_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/renstrom/dedent"

	"regexp"

	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks"
	stubs "gitlab.com/mr-meeseeks/meeseeks-box/testingstubs"
)

type expectedMessage struct {
	TextMatcher string
	Channel     string
	IsIM        bool
}

func Test_BasicReplying(t *testing.T) {
	handshakeMatcher := fmt.Sprintf("^(%s)$", strings.Join(config.DefaultHandshake, "|"))

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
					Channel:     "general",
					IsIM:        false,
				},
				expectedMessage{
					TextMatcher: "^<@myuser> .*\n```\nhello!\n```$",
					Channel:     "general",
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
					Channel:     "general",
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
					Channel:     "general",
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
					TextMatcher: "<@myuser> Uuuuh, yeah! you are not allowed to do false",
					Channel:     "general",
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
					Channel:     "general",
					IsIM:        false,
				},
				expectedMessage{
					TextMatcher: "^<@myuser> Uuuh!, no, it failed :disappointed: exit status 1$",
					Channel:     "general",
					IsIM:        false,
				},
			},
		},
	}

	client, cnf := stubs.NewHarness().
		WithConfig(dedent.Dedent(`
			---
			commands:
			  fail:
			    command: false
			    timeout: 10
			    auth_strategy: any
			  disallowed:
			    command: false
			    timeout: 10
			    auth_strategy: none
			`)).Build()

	m := meeseeks.New(client, cnf)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			go m.Process(stubs.MessageStub{
				Text:    tc.message,
				Channel: tc.channel,
				User:    tc.user,
			})

			for _, expected := range tc.expected {
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

}
