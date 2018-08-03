package formatter

import (
	"strings"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/text/template"

	"github.com/sirupsen/logrus"
)

// Default colors
const (
	DefaultInfoColorMessage    = ""
	DefaultSuccessColorMessage = "good"
	DefaultWarningColorMessage = "warning"
	DefaultErrColorMessage     = "danger"
)

// MessageColors contains the configured reply message colora
type MessageColors struct {
	Info    string `yaml:"info"`
	Success string `yaml:"success"`
	Error   string `yaml:"error"`
}

// FormatConfig contains the formatting configurations
type FormatConfig struct {
	Colors     MessageColors       `yaml:"colors"`
	ReplyStyle map[string]string   `yaml:"reply_styles"`
	Templates  map[string]string   `yaml:"templates"`
	Messages   map[string][]string `yaml:"messages"`
}

// Formatter keeps the colors and templates used to format a reply message
type Formatter struct {
	colors     MessageColors
	templates  *template.TemplatesBuilder
	replyStyle replyStyle
}

var formatter *Formatter

// Configure sets up the singleton formatter
func Configure(cnf FormatConfig) {
	builder := template.NewBuilder().WithMessages(cnf.Messages).WithTemplates(cnf.Templates)
	formatter = &Formatter{
		replyStyle: replyStyle{cnf.ReplyStyle},
		colors:     cnf.Colors,
		templates:  builder,
	}
}

// Templates returns a clone of the default templates ready to be consumed
func Templates() template.Templates {
	return formatter.templates.Clone().Build()
}

// WithTemplates returns a clone of the default templates with the templates
// passed as argument applied on top
func WithTemplates(templates map[string]string) template.Templates {
	return formatter.templates.Clone().WithTemplates(templates).Build()
}

// HandshakeReply creates a reply for a handshake message
func HandshakeReply(req meeseeks.Request) Reply {
	return formatter.newReplier(template.Handshake, req)
}

// UnknownCommandReply creates a reply for an UnknownCommand error message
func UnknownCommandReply(req meeseeks.Request) Reply {
	return formatter.newReplier(template.UnknownCommand, req)
}

// UnauthorizedCommandReply creates a reply for an unauthorized command error message
func UnauthorizedCommandReply(req meeseeks.Request) Reply {
	return formatter.newReplier(template.Unauthorized, req)
}

// FailureReply creates a reply for a generic command error message
func FailureReply(req meeseeks.Request, err error) Reply {
	return formatter.newReplier(template.Failure, req).WithError(err)
}

// SuccessReply creates a reply for a generic command success message
func SuccessReply(req meeseeks.Request) Reply {
	return formatter.newReplier(template.Success, req)
}

func (f Formatter) newReplier(action string, req meeseeks.Request) Reply {
	style := f.replyStyle.Get(action)
	logrus.Debugf("creating replier '%s' for action %s", style, action)

	return Reply{
		action:  action,
		request: req,

		templates: f.templates.Clone(),
		style:     style,
		colors:    f.colors,
	}
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
	action  string
	request meeseeks.Request
	output  string
	err     error

	colors    MessageColors
	templates *template.TemplatesBuilder
	style     string
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
	payload := make(map[string]interface{})
	payload["command"] = r.request.Command
	payload["args"] = strings.Join(r.request.Args, " ")

	payload["user"] = r.request.Username
	payload["userlink"] = r.request.UserLink
	payload["userid"] = r.request.UserID
	payload["channel"] = r.request.Channel
	payload["channellink"] = r.request.ChannelLink
	payload["channelid"] = r.request.ChannelID
	payload["isim"] = r.request.IsIM

	payload["error"] = r.err
	payload["output"] = r.output

	return r.templates.Build().Render(r.action, payload)
}

// ChannelID returns the channel ID in which to reply
func (r Reply) ChannelID() string {
	return r.request.ChannelID
}

// ReplyStyle returns the style to use to reply
func (r Reply) ReplyStyle() string {
	return r.style
}

// Color returns the color to use when decorating the reply
func (r Reply) Color() string {
	switch r.action {
	case template.Handshake:
		return r.colors.Info
	case template.UnknownCommand, template.Unauthorized, template.Failure:
		return r.colors.Error
	default:
		return r.colors.Success
	}
}
