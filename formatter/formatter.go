package formatter

import (
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

// ErrorColor returns the color used to represent an error
func (f Formatter) ErrorColor() string {
	return f.colors.Error
}

// InfoColor returns the color used to represent an informational message
func (f Formatter) InfoColor() string {
	return f.colors.Info
}

// SuccessColor returns the color used to represent a successful command
func (f Formatter) SuccessColor() string {
	return f.colors.Success
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
