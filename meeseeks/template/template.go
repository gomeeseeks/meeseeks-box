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
	defaultHandshakeTemplate = "{{ .AnyMessage \"Handshake\" }}"
	defaultSuccessTemplate   = "{{ .User }} {{ .AnyMessage \"Success\" }}{{ with .Output }}\n\nOutput:\n```\n{{ .Output }}```{{ end }}"
	defaultFailureTemplate   = "{{ .User }} {{ .AnyMessage \"Failed\" }} :disappointed:: {{ .Error }}{{ with .Output }}\n\nOutput:\n```\n{{ .Output }}```{{ end }}"
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
	Handshake ReplyTemplate
	Success   ReplyTemplate
	Failure   ReplyTemplate
}

// TemplateData is a helper type that provides a AnyMessage(key) method
type TemplateData map[string]interface{}

// AnyMessage picks a random string from the list of strings contained in `key`
func (p TemplateData) AnyMessage(key string) string {
	messages, ok := p[key].([]string)
	if !ok {
		return fmt.Sprintf("ERROR: %s is not a string slice", key)
	}
	return messages[rand.Intn(len(messages))]
}

// ReplyTemplate is a pre rendered template used to reply
type ReplyTemplate struct {
	template *tmpl.Template
}

// New creates a new ReplyTemplate pre-parsing the template
func New(name, template string) (ReplyTemplate, error) {
	t, err := tmpl.New(name).Parse(template)
	if err != nil {
		return ReplyTemplate{}, fmt.Errorf("could not parse template %s: %s", name, err)
	}
	return ReplyTemplate{
		template: t,
	}, nil
}

// Render renders the template with the passed in data
func (r ReplyTemplate) Render(data interface{}) (string, error) {
	b := bytes.NewBuffer([]byte{})
	err := r.template.Execute(b, data)

	if err != nil {
		return "", fmt.Errorf("failed to execute template %s: %s", r.template.Name, err)
	}

	return string(b.Bytes()), nil
}
