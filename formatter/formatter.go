package formatter

import (
	"github.com/pcarranza/meeseeks-box/config"
	"github.com/pcarranza/meeseeks-box/template"
)

type Formatter struct {
	colors    config.MessageColors
	templates *template.TemplatesBuilder
}

func New(cnf config.Config) *Formatter {
	builder := template.NewBuilder().WithMessages(cnf.Messages)
	return &Formatter{
		colors:    cnf.Colors,
		templates: builder,
	}
}

func (f Formatter) ErrorColor() string {
	return f.colors.Error
}

func (f Formatter) InfoColor() string {
	return f.colors.Info
}

func (f Formatter) SuccessColor() string {
	return f.colors.Success
}

func (f Formatter) Templates() template.Templates {
	return f.templates.Clone().Build()
}

func (f Formatter) WithTemplates(templates map[string]string) template.Templates {
	return f.templates.Clone().WithTemplates(templates).Build()
}
