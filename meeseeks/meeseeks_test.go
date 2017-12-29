package meeseeks_test

import (
	"bytes"
	"fmt"
	"testing"

	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks"
)

func Test_ReplyingWithEcho(t *testing.T) {
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
	client := NewClientStub()
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			m := MessageStub{
				text:    tc.message,
				channel: tc.channel,
				user:    tc.user,
			}
			meeseeks.ProcessMessage(m, &client)
			if !client.Contains(tc.expected) {
				t.Fatalf("can't find message %s; got %s", tc.expected, client)
			}
		})
	}

}

type SentMessage struct {
	text    string
	channel string
	im      bool
}

func (s SentMessage) String() string {
	return fmt.Sprintf("channel: %s text: %s im: %t", s.channel, s.text, s.im)
}

type ClientStub struct {
	messages []SentMessage
}

func NewClientStub() ClientStub {
	return ClientStub{
		messages: make([]SentMessage, 0),
	}
}

func (c *ClientStub) Reply(text, channel string) {
	c.messages = append(c.messages, SentMessage{text: text, channel: channel})
}

func (c *ClientStub) ReplyIM(text, user string) error {
	c.messages = append(c.messages, SentMessage{text: text, channel: user, im: true})
	return nil
}

func (c ClientStub) String() string {
	b := bytes.NewBufferString("")
	for _, m := range c.messages {
		b.WriteString(fmt.Sprintf("%s\n", m))
	}
	return b.String()
}

func (c ClientStub) Contains(message string) bool {
	for _, m := range c.messages {
		if message == fmt.Sprint(m) {
			return true
		}
	}
	return false
}

type MessageStub struct {
	text    string
	channel string
	user    string
}

func (m MessageStub) GetText() string {
	return m.text
}

func (m MessageStub) GetChannel() string {
	return m.channel
}

func (m MessageStub) GetUserFrom() string {
	return fmt.Sprintf("<@%s>", m.user)
}
