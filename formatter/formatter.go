package formatter

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/gomeeseeks/meeseeks-box/config"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/template"
)

// Formatter keeps the colors and templates used to format a reply message
type Formatter struct {
	colors     config.MessageColors
	templates  *template.TemplatesBuilder
	replyStyle replyStyle
}

// New returns a new Formatter
func New(cnf config.Config) *Formatter {
	builder := template.NewBuilder().WithMessages(cnf.Messages)
	f := Formatter{
		replyStyle: replyStyle{cnf.Format.ReplyStyle},
		colors:     cnf.Format.Colors,
		templates:  builder,
	}
	logrus.Debugf("Building new formatter %#v", f)
	return &f
}

// Templates returns a clone of the default templates ready to be consumed
func (f Formatter) Templates() template.Templates {
	return f.templates.Clone().Build()
}

// WithTemplates returns a clone of the default templates with the templates
// passed as argument applied on top
func (f Formatter) WithTemplates(templates map[string]string) template.Templates {
	return f.templates.Clone().WithTemplates(templates).Build()
}

// HandshakeReply creates a reply for a handshake message
func (f Formatter) HandshakeReply(to ReplyTo) Reply {
	return f.newReplier(template.Handshake, to)
}

// UnknownCommandReply creates a reply for an UnknownCommand error message
func (f Formatter) UnknownCommandReply(to ReplyTo, cmd string) Reply {
	return f.newReplier(template.UnknownCommand, to).WithOutput(cmd)
}

// UnauthorizedCommandReply creates a reply for an unauthorized command error message
func (f Formatter) UnauthorizedCommandReply(to ReplyTo, cmd string) Reply {
	return f.newReplier(template.Unauthorized, to).WithOutput(cmd)
}

// FailureReply creates a reply for a generic command error message
func (f Formatter) FailureReply(to ReplyTo, err error) Reply {
	return f.newReplier(template.Failure, to).WithError(err)
}

// SuccessReply creates a reply for a generic command success message
func (f Formatter) SuccessReply(to ReplyTo) Reply {
	return f.newReplier(template.Success, to)
}

func (f Formatter) newReplier(mode string, to ReplyTo) Reply {
	return Reply{
		mode: mode,
		to:   to,

		templates: f.templates.Clone(),
		style:     f.replyStyle.Get(mode),
		colors:    f.colors,
	}
}

// ReplyTo holds the information of who and where to reply
type ReplyTo struct {
	UserLink  string
	ChannelID string
}

type replyStyle struct {
	styles map[string]string
}

func (r replyStyle) Get(mode string) string {
	switch mode {
	case template.Handshake,
		template.UnknownCommand,
		template.Unauthorized,
		template.Failure,
		template.Success:

		if style, ok := r.styles[mode]; ok {
			return style
		}
	}
	return ""
}

// Reply represents all the data necessary to send a reply message
type Reply struct {
	mode   string
	to     ReplyTo
	output string
	err    error

	colors    config.MessageColors
	templates *template.TemplatesBuilder
	style     string
}

// WithCommand receives a command and pulls all the specific command configuration
func (r Reply) WithCommand(cmd meeseeks.Command) Reply {
	if r.templates == nil {
		logrus.Info("templates are nil in the current reply??")
		return r
	}
	if cmd.Templates() == nil {
		logrus.Info("command templates are nil for %#v", cmd)
		return r
	}

	r.templates = r.templates.WithTemplates(cmd.Templates())
	return r
}

// WithOutput stores the text payload to render in the reply
func (r Reply) WithOutput(output string) Reply {
	r.output = output
	return r
}

// WithError stores an error to render
func (r Reply) WithError(err error) Reply {
	r.err = err
	return r
}

// Render renders the message returning the rendered text, or an error if something goes wrong.
func (r Reply) Render() (string, error) {
	switch r.mode {
	case template.Handshake:
		return r.templates.Build().RenderHandshake(r.to.UserLink)
	case template.UnknownCommand:
		return r.templates.Build().RenderUnknownCommand(r.to.UserLink, r.output)
	case template.Unauthorized:
		return r.templates.Build().RenderUnauthorizedCommand(r.to.UserLink, r.output, r.err)
	case template.Failure:
		return r.templates.Build().RenderFailure(r.to.UserLink, r.err.Error(), r.output)
	case template.Success:
		return r.templates.Build().RenderSuccess(r.to.UserLink, r.output)
	default:
		return "", fmt.Errorf("don't know how to render mode '%s'", r.mode)
	}
}

// ChannelID returns the channel ID in which to reply
func (r Reply) ChannelID() string {
	return r.to.ChannelID
}

// ReplyStyle returns the style to use to reply
func (r Reply) ReplyStyle() string {
	return r.style
}

// Color returns the color to use when decorating the reply
func (r Reply) Color() string {
	switch r.mode {
	case template.Handshake:
		return r.colors.Info
	case template.UnknownCommand, template.Unauthorized, template.Failure:
		return r.colors.Error
	default:
		return r.colors.Success
	}
}
