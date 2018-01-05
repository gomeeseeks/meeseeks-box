package commands

import (
	"fmt"

	"github.com/renstrom/dedent"
	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/template"
	"gitlab.com/mr-meeseeks/meeseeks-box/version"
)

type builtinCommand struct {
}

var builtinTemplates = map[string]string{
	template.SuccessKey: fmt.Sprintf("{{ .user }} {{ AnyValue \"%s\" . }}{{ with $out := .output }}\n{{ $out }}{{ end }}", template.SuccessKey),
}

var allowAllConfiguredCommand = config.Command{
	AuthStrategy: config.AuthStrategyAny,
	Templates:    builtinTemplates,
}

func (b builtinCommand) HasHandshake() bool {
	return false
}

func (b builtinCommand) ConfiguredCommand() config.Command {
	return allowAllConfiguredCommand
}

type versionCommand struct {
	builtinCommand
	Help string
}

func (v versionCommand) Execute(args ...string) (string, error) {
	return version.AppVersion, nil
}

type helpCommand struct {
	builtinCommand
	commands *map[string]Command
	Help     string
}

func (h helpCommand) Execute(args ...string) (string, error) {
	tmpl, err := template.New("version", dedent.Dedent(
		`{{ range $name, $cmd := .commands }}
		- {{ $name }}: {{ $cmd.Help }}
		{{- end }}
		`))
	if err != nil {
		return "", err
	}
	return tmpl.Render(template.Payload{
		"commands": h.commands,
	})
}
