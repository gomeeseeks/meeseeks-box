package formatter

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/gomeeseeks/meeseeks-box/command"
	"github.com/gomeeseeks/meeseeks-box/config"
	"github.com/gomeeseeks/meeseeks-box/template"
)

// Formatter keeps the colors and templates used to format a reply message
type Formatter struct {
	colors     config.MessageColors
	templates  *template.TemplatesBuilder
	replyStyle ReplyStyle
}

// New returns a new Formatter
func New(cnf config.Config) *Formatter {
	builder := template.NewBuilder().WithMessages(cnf.Messages)
	f := Formatter{
		replyStyle: ReplyStyle{cnf.Format.ReplyStyle},
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

func (f Formatter) HandshakeReply(to ReplyTo) Reply {
	return f.newReplier(template.Handshake, to)
}

func (f Formatter) UnknownCommandReply(to ReplyTo, cmd string) Reply {
	return f.newReplier(template.UnknownCommand, to).WithOutput(cmd)
}

func (f Formatter) UnauthorizedCommandReply(to ReplyTo, cmd string) Reply {
	return f.newReplier(template.Unauthorized, to).WithOutput(cmd)
}

func (f Formatter) FailureReply(to ReplyTo, err error) Reply {
	return f.newReplier(template.Failure, to).WithError(err)
}

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

type ReplyTo struct {
	UserLink  string
	ChannelID string
}

type ReplyStyle struct {
	styles map[string]string
}

func (r ReplyStyle) Get(mode string) string {
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

type Reply struct {
	mode    string
	to      ReplyTo
	command string
	output  string
	err     error

	colors    config.MessageColors
	templates *template.TemplatesBuilder
	style     string
}

func (r Reply) WithCommand(cmd command.Command) Reply {
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

func (r Reply) WithOutput(output string) Reply {
	r.output = output
	return r
}

func (r Reply) WithError(err error) Reply {
	r.err = err
	return r
}

func (r Reply) WithStyle(style string) Reply {
	r.style = style
	return r
}

func (r Reply) Render() (string, error) {
	switch r.mode {
	case template.Handshake:
		return r.templates.Build().RenderHandshake(r.to.UserLink)
	case template.UnknownCommand:
		return r.templates.Build().RenderUnknownCommand(r.to.UserLink, r.output)
	case template.Unauthorized:
		return r.templates.Build().RenderUnauthorizedCommand(r.to.UserLink, r.output)
	case template.Failure:
		return r.templates.Build().RenderFailure(r.to.UserLink, r.err.Error(), r.output)
	case template.Success:
		return r.templates.Build().RenderSuccess(r.to.UserLink, r.output)
	default:
		return "", fmt.Errorf("Don't know how to render mode '%s'", r.mode)
	}
}

func (r Reply) Channel() string {
	return r.to.ChannelID
}

func (r Reply) ReplyStyle() string {
	return r.style
}

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
