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
	})

	t.Run("failure", func(t *testing.T) {
		r := formatter.FailureReply(meeseeks.Request{
			Command: "test",
		}, errors.New("some error"))
		s, e := r.Render()
		mocks.Must(t, "could not render error reply", e)

		mocks.AssertEquals(t, "test failure! some error", s)
		mocks.AssertEquals(t, "", r.ReplyStyle())

	})

	t.Run("success with output", func(t *testing.T) {
		r := formatter.SuccessReply(meeseeks.Request{
			Command: "test",
		})
		s, e := r.WithOutput(" output!").Render()
		mocks.Must(t, "could not render successful reply with output", e)

		mocks.AssertEquals(t, "test success! output!", s)
		mocks.AssertEquals(t, "text", r.ReplyStyle())

	})
	tt := []struct {
		name          string
		f             func(meeseeks.Request) formatter.Reply
		expectedText  string
		expectedStyle string
	}{
		{
			name:          template.Handshake,
			f:             formatter.HandshakeReply,
			expectedText:  "test hello",
			expectedStyle: "text",
		}, {
			name:          template.Success,
			f:             formatter.SuccessReply,
			expectedText:  "test success!",
			expectedStyle: "text",
		}, {
			name:          template.Unauthorized,
			f:             formatter.UnauthorizedCommandReply,
			expectedText:  "test unauthorized!",
			expectedStyle: "attachment",
		}, {
			name:          template.UnknownCommand,
			f:             formatter.UnknownCommandReply,
			expectedText:  "test unknown!",
			expectedStyle: "attachment",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.f(meeseeks.Request{
				Command: "test",
			})
			s, e := r.Render()
			mocks.Must(t, "could not render reply", e)

			mocks.AssertEquals(t, tc.expectedText, s)
			mocks.AssertEquals(t, tc.expectedStyle, r.ReplyStyle())
		})
	}
}
