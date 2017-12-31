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
		"{{ with $out := .output }}\n\nOutput:\n```\n{{ $out }}```{{ end }}"
	defaultFailureTemplate = "{{ .user }} {{ AnyValue \"failed\" . }} :disappointed:: {{ .error }}" +
		"{{ with $out := .output }}\n\nOutput:\n```\n{{ $out }}```{{ end }}"
)

// DefaultTemplates builds a set of default template renderers
func DefaultTemplates() Templates {
	handshake, err := New("handshake", defaultHandshakeTemplate)
	if err != nil {
		log.Fatalf("could not parse default handshake template: %s", err)
	}

	success, err := New("success", defaultSuccessTemplate)
	if err != nil {
		log.Fatalf("could not parse default success template: %s", err)
	}

	failure, err := New("failure", defaultFailureTemplate)
	if err != nil {
		log.Fatalf("could not parse default failure template: %s", err)
	}

	return Templates{
		Handshake: handshake,
		Success:   success,
		Failure:   failure,
	}
}

// Templates is a set of templates for the basic operations
type Templates struct {
	Handshake Renderer
	Success   Renderer
	Failure   Renderer
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
