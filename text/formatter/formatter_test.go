package formatter_test

import (
	"errors"
	"testing"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/mocks"
	"github.com/gomeeseeks/meeseeks-box/text/formatter"
	"github.com/gomeeseeks/meeseeks-box/text/template"
)

func TestFormatterTemplates(t *testing.T) {
	formatter.Configure(formatter.FormatConfig{
		Templates: map[string]string{
			template.Handshake:      "{{ .command }} hello",
			template.Success:        "{{ .command }} success!{{ .output }}",
			template.Failure:        "{{ .command }} failure! {{ .error }}",
			template.Unauthorized:   "{{ .command }} unauthorized!",
			template.UnknownCommand: "{{ .command }} unknown!",
		},
		ReplyStyle: map[string]string{
			template.Handshake:      "text",
			template.Success:        "text",
			template.Unauthorized:   "attachment",
			template.UnknownCommand: "attachment",
		},
		Colors: formatter.MessageColors{
			Success: "green",
			Error:   "red",
			Info:    "blue",
		},
	})

	t.Run("failure", func(t *testing.T) {
		r := formatter.FailureReply(meeseeks.Request{
			Command:   "test",
			ChannelID: "failure",
		}, errors.New("some error"))
		s, e := r.Render()
		mocks.Must(t, "could not render error reply", e)

		mocks.AssertEquals(t, "test failure! some error", s)
		mocks.AssertEquals(t, "", r.ReplyStyle())
		mocks.AssertEquals(t, "failure", r.ChannelID())
		mocks.AssertEquals(t, "red", r.Color())

	})

	t.Run("success with output", func(t *testing.T) {
		r := formatter.SuccessReply(meeseeks.Request{
			Command:   "test",
			ChannelID: "success-with-output",
		})
		s, e := r.WithOutput(" output!").Render()
		mocks.Must(t, "could not render successful reply with output", e)

		mocks.AssertEquals(t, "test success! output!", s)
		mocks.AssertEquals(t, "text", r.ReplyStyle())
		mocks.AssertEquals(t, "success-with-output", r.ChannelID())
		mocks.AssertEquals(t, "green", r.Color())

	})

	tt := []struct {
		name          string
		f             func(meeseeks.Request) formatter.Reply
		expectedText  string
		expectedStyle string
		expectedColor string
	}{
		{
			name:          template.Handshake,
			f:             formatter.HandshakeReply,
			expectedText:  "test hello",
			expectedStyle: "text",
			expectedColor: "blue",
		}, {
			name:          template.Success,
			f:             formatter.SuccessReply,
			expectedText:  "test success!",
			expectedStyle: "text",
			expectedColor: "green",
		}, {
			name:          template.Unauthorized,
			f:             formatter.UnauthorizedCommandReply,
			expectedText:  "test unauthorized!",
			expectedStyle: "attachment",
			expectedColor: "red",
		}, {
			name:          template.UnknownCommand,
			f:             formatter.UnknownCommandReply,
			expectedText:  "test unknown!",
			expectedStyle: "attachment",
			expectedColor: "red",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.f(meeseeks.Request{
				Command:   "test",
				ChannelID: tc.name + "channel",
			})
			s, e := r.Render()
			mocks.Must(t, "could not render reply", e)

			mocks.AssertEquals(t, tc.expectedText, s)
			mocks.AssertEquals(t, tc.expectedStyle, r.ReplyStyle())
			mocks.AssertEquals(t, tc.name+"channel", r.ChannelID())
			mocks.AssertEquals(t, tc.expectedColor, r.Color())
		})
	}
}
