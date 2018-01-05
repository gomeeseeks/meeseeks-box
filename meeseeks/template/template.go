package template

import (
	"bytes"
	"fmt"
	"math/rand"
	tmpl "text/template"

	log "github.com/sirupsen/logrus"
)

// Default command templates
const (
	defaultHandshakeTemplate = "{{ AnyValue \"handshake\" . }}"
	defaultSuccessTemplate   = "{{ .user }} {{ AnyValue \"success\" . }}" +
		"{{ with $out := .output }}\n```\n{{ $out }}```{{ end }}"
	defaultFailureTemplate = "{{ .user }} {{ AnyValue \"failed\" . }} :disappointed: {{ .error }}" +
		"{{ with $out := .output }}\n```\n{{ $out }}```{{ end }}"
	defaultUnknownCommandTemplate = "{{ .user }} {{ AnyValue \"unknowncommand\" . }} {{ .command }}"
	defaultUnauthorizedTemplate   = "{{ .user }} {{ AnyValue \"unauthorized\" . }} {{ .command }}"
)

// Default messages
var (
	DefaultHandshakeMessages      = []string{"I'm Mr Meeseeks! look at me!", "Mr Meeseeks!", "Uuuuh, yeah! can do!", "Can doo...", "Uuuuh, ok!"}
	DefaultSuccessMessages        = []string{"All done!", "Mr Meeseeks", "Uuuuh, nice!"}
	DefaultFailedMessages         = []string{"Uuuh!, no, it failed"}
	DefaultUnauthorizedMessages   = []string{"Uuuuh, yeah! you are not allowed to do"}
	DefaultUnknownCommandMessages = []string{"Uuuh! no, I don't know how to do"}
)

// Template names used for rendering
const (
	HandshakeTemplate      = "handshake"
	SuccessTemplate        = "success"
	FailureTemplate        = "failure"
	UnknownCommandTemplate = "unknowncommand"
	UnauthorizedTemplate   = "unauthorized"
)

// Templates is a set of templates for the basic operations
type Templates struct {
	renderers map[string]Renderer
	// Handshake      Renderer
	// Success        Renderer
	// Failure        Renderer
	// UnknownCommand Renderer
	// Unauthorized   Renderer
	defaultPayload Payload
}

type TemplatesBuilder struct {
	messages  map[string][]string
	templates map[string]string
}

func NewBuilder() TemplatesBuilder {
	return TemplatesBuilder{
		templates: map[string]string{
			"handshake":      defaultHandshakeTemplate,
			"success":        defaultSuccessTemplate,
			"failure":        defaultFailureTemplate,
			"unknowncommand": defaultUnknownCommandTemplate,
			"unauthorized":   defaultUnauthorizedTemplate,
		},
		messages: map[string][]string{
			"handshake":      DefaultHandshakeMessages,
			"success":        DefaultSuccessMessages,
			"failed":         DefaultFailedMessages,
			"unknowncommand": DefaultUnknownCommandMessages,
			"unauthorized":   DefaultUnauthorizedMessages,
		},
	}
}

func (b TemplatesBuilder) WithMessages(messages map[string][]string) TemplatesBuilder {
	for name, message := range messages {
		b.messages[name] = message
	}
	return b
}

func (b TemplatesBuilder) Build() Templates {
	renderers := make(map[string]Renderer)
	for name, template := range b.templates {
		renderer, err := New(name, template)
		if err != nil {
			log.Fatalf("could not parse %s template: %s", name, err)
		}
		renderers[name] = renderer
	}

	payload := Payload{}
	for k, v := range b.messages {
		payload[k] = v
	}

	return Templates{
		renderers:      renderers,
		defaultPayload: payload,
	}
}

// RenderHandshake renders a handshake message
func (t Templates) RenderHandshake(user string) (string, error) {
	p := t.newPayload()
	p["user"] = user
	return t.renderers[HandshakeTemplate].Render(p)
}

// RenderUnknownCommand renders an unknown command message
func (t Templates) RenderUnknownCommand(user, cmd string) (string, error) {
	p := t.newPayload()
	p["user"] = user
	p["command"] = cmd
	return t.renderers[UnknownCommandTemplate].Render(p)
}

// RenderUnauthorizedCommand renders an unauthorized command message
func (t Templates) RenderUnauthorizedCommand(user, cmd string) (string, error) {
	p := t.newPayload()
	p["user"] = user
	p["command"] = cmd
	return t.renderers[UnauthorizedTemplate].Render(p)
}

// RenderSuccess renders a success message
func (t Templates) RenderSuccess(user, output string) (string, error) {
	p := t.newPayload()
	p["user"] = user
	p["output"] = output
	return t.renderers[SuccessTemplate].Render(p)
}

// RenderFailure renders a failure message
func (t Templates) RenderFailure(user, err, output string) (string, error) {
	p := t.newPayload()
	p["user"] = user
	p["error"] = err
	p["output"] = output
	return t.renderers[FailureTemplate].Render(p)
}

func (t Templates) newPayload() Payload {
	p := Payload{}
	for k, v := range t.defaultPayload {
		p[k] = v
	}
	return p
}

// Payload is a helper type that provides a AnyMessage(key) method
type Payload map[string]interface{}

func anyValue(key string, payload map[string]interface{}) (string, error) {
	values, ok := payload[key]
	if !ok {
		return "", fmt.Errorf("ERROR: %s is not loaded in the payload", key)
	}
	slice, ok := values.([]string)
	if !ok {
		return "", fmt.Errorf("ERROR: %s is not a string slice", key)
	}
	return slice[rand.Intn(len(slice))], nil
}

// Renderer is a pre rendered template used to reply
type Renderer struct {
	template *tmpl.Template
}

// New creates a new ReplyTemplate pre-parsing the template
func New(name, template string) (Renderer, error) {
	t, err := tmpl.New(name).Funcs(tmpl.FuncMap{
		"AnyValue": anyValue,
	}).Parse(template)
	if err != nil {
		return Renderer{}, fmt.Errorf("could not parse template %s: %s", name, err)
	}
	return Renderer{
		template: t,
	}, nil
}

// Render renders the template with the passed in data
func (r Renderer) Render(data Payload) (string, error) {
	b := bytes.NewBuffer([]byte{})
	err := r.template.Execute(b, data)

	if err != nil {
		return "", fmt.Errorf("failed to execute template %s: %s", r.template.Name(), err)
	}

	return string(b.Bytes()), nil
}
