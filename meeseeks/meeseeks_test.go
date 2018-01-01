package meeseeks_test

import (
	"testing"

	"regexp"

	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks"
	stubs "gitlab.com/mr-meeseeks/meeseeks-box/testingstubs"
)

func Test_BasicReplying(t *testing.T) {
	tt := []struct {
		name              string
		user              string
		message           string
		channel           string
		expectedTextMatch string
		expectedChannel   string
		expectedIM        bool
	}{
		{
			name:              "basic case",
			user:              "myuser",
			message:           "echo hello!",
			channel:           "general",
			expectedTextMatch: "^<@myuser> .*\n\nOutput:\n```\nhello!\n```$",
			expectedChannel:   "general",
			expectedIM:        false,
		},
		{
			name:              "unknown command case",
			user:              "myuser",
			message:           "unknown-command hello!",
			channel:           "general",
			expectedTextMatch: "^<@myuser> Uuuh! no, I don't know how to do unknown-command$",
			expectedChannel:   "general",
			expectedIM:        false,
		},
	}

	client, cnf := stubs.NewHarness().Build()
	m := meeseeks.New(client, cnf)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			go m.Process(stubs.MessageStub{
				Text:    tc.message,
				Channel: tc.channel,
				User:    tc.user,
			})

			actual := <-client.Messages

			r, err := regexp.Compile(tc.expectedTextMatch)
			stubs.Must(t, "could not compile regex", err, tc.expectedTextMatch)

			if !r.MatchString(actual.Text) {
				t.Fatalf("Bad message, expected %s; got %s", tc.expectedTextMatch, actual.Text)
			}
			stubs.AssertEquals(t, tc.expectedChannel, actual.Channel)
			stubs.AssertEquals(t, tc.expectedIM, actual.Im)
		})
	}

}
