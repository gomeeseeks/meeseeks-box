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
			name:     "basic case",
			user:     "myuser",
			message:  "hello!",
			channel:  "general",
			expected: "channel: general text: <@myuser> echo: hello! im: false",
		},
	}
	client := stubs.NewClientStub()
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			m := stubs.MessageStub{
				Text:    tc.message,
				Channel: tc.channel,
				User:    tc.user,
			}
			meeseeks.ProcessMessage(m, &client)
			if !client.Contains(tc.expected) {
				t.Fatalf("can't find message %s; got %s", tc.expected, client)
			}
		})
	}

}
