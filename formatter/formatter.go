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
	colors    config.MessageColors
	templates *template.TemplatesBuilder
}

// New returns a new Formatter
func New(cnf config.Config) *Formatter {
	builder := template.NewBuilder().WithMessages(cnf.Messages)
	return &Formatter{
		colors:    cnf.Colors,
		templates: builder,
	}
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

const (
	Handshake           = "handshake"
	UnknownCommand      = "unknown"
	UnauthorizedCommand = "unauthorized"
	Error               = "error"
	Success             = "success"
)

func (f Formatter) HandshakeReply(to ReplyTo) Reply {
	return f.newReplier(Handshake, to)
}

func (f Formatter) UnknownCommandReply(to ReplyTo, cmd string) Reply {
	return f.newReplier(UnknownCommand, to).WithOutput(cmd)
}

func (f Formatter) UnauthorizedCommandReply(to ReplyTo, cmd string) Reply {
	return f.newReplier(UnauthorizedCommand, to).WithOutput(cmd)
}

func (f Formatter) ErrorReply(to ReplyTo, err error) Reply {
	return f.newReplier(Error, to).WithError(err)
}

func (f Formatter) SuccessReply(to ReplyTo) Reply {
	return f.newReplier(Success, to)
}

func (f Formatter) newReplier(kind string, to ReplyTo) Reply {
	tmpls := f.templates.Clone()
	return Reply{
		kind:      kind,
		to:        to,
		templates: tmpls,
	}
}

type ReplyTo struct {
	UserLink  string
	ChannelID string
}

type Reply struct {
	kind    string
	to      ReplyTo
	command string
	output  string
	err     error

	colors    config.MessageColors
	templates *template.TemplatesBuilder
}

// func (r Reply) Output() string {
// 	return r.output
// }

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

func (r Reply) Render() (string, error) {
	switch r.kind {
	case Handshake:
		return r.templates.Build().RenderHandshake(r.to.UserLink)
	case UnknownCommand:
		return r.templates.Build().RenderUnknownCommand(r.to.UserLink, r.output)
	case UnauthorizedCommand:
		return r.templates.Build().RenderUnauthorizedCommand(r.to.UserLink, r.output)
	case Error:
		return r.templates.Build().RenderFailure(r.to.UserLink, r.err.Error(), r.output)
	case Success:
		return r.templates.Build().RenderSuccess(r.to.UserLink, r.output)
	default:
		return "", fmt.Errorf("Don't know how to render kind '%s'", r.kind)
	}
}

func (r Reply) Channel() string {
	return r.to.ChannelID
}

func (r Reply) Kind() string {
	return r.kind
}

func (r Reply) Color() string {
	switch r.kind {
	case Handshake:
		return r.colors.Info
	case UnknownCommand, UnauthorizedCommand, Error:
		return r.colors.Error
	default:
		return r.colors.Success
	}
}
