package meeseeks_test

import (
	"testing"

	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks"
	stubs "gitlab.com/mr-meeseeks/meeseeks-box/testingstubs"
)

func Test_BasicReplying(t *testing.T) {
	tt := []struct {
		name     string
		user     string
		message  string
		channel  string
		expected string
	}{
		{
			name:    "basic case",
			user:    "myuser",
			message: "echo hello!",
			channel: "general",
			// expected: "channel: general text: <@myuser> Done! \nOutput:```hello!``` im: false",
			expected: "channel: general text: <@myuser> echo: echo hello! im: false",
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

			if tc.expected != actual.String() {
				t.Fatalf("can't find message %s; got %s", tc.expected, actual)
			}
		})
	}

}
