package command

import (
	"flag"
	"fmt"
	"strings"

	"gitlab.com/mr-meeseeks/meeseeks-box/jobs"

	"github.com/renstrom/dedent"
	"gitlab.com/mr-meeseeks/meeseeks-box/auth"
	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/request"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/template"
	"gitlab.com/mr-meeseeks/meeseeks-box/version"
)

var builtinTemplates = map[string]string{
	template.SuccessKey: fmt.Sprintf("{{ .user }} {{ AnyValue \"%s\" . }}{{ with $out := .output }}\n{{ $out }}{{ end }}", template.SuccessKey),
}

var allowAllCommand = config.Command{
	AuthStrategy: config.AuthStrategyAny,
	Templates:    builtinTemplates,
}

var allowAdminsCommand = config.Command{
	AuthStrategy:  config.AuthStrategyAllowedGroup,
	Templates:     builtinTemplates,
	AllowedGroups: []string{config.AdminGroup},
}

type noHandshake struct {
}

func (b noHandshake) HasHandshake() bool {
	return false
}

type allowAll struct {
}

func (a allowAll) ConfiguredCommand() config.Command {
	return allowAllCommand
}

type allowAdmins struct {
}

func (a allowAdmins) ConfiguredCommand() config.Command {
	return allowAdminsCommand
}

type versionCommand struct {
	noHandshake
	allowAll
	Help string
}

func (v versionCommand) Execute(req request.Request) (string, error) {
	return version.Version, nil
}

type helpCommand struct {
	noHandshake
	allowAll
	commands *map[string]Command
	Help     string
}

func (h helpCommand) Execute(req request.Request) (string, error) {
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

type groupsCommand struct {
	noHandshake
	allowAdmins
	Help string
}

func (g groupsCommand) Execute(req request.Request) (string, error) {
	tmpl, err := template.New("version", dedent.Dedent(`
		{{- range $group, $users := .groups }}
		- {{ $group }}:
		  {{- range $index, $user := $users }}{{ if ne $index 0 }},{{ end }} {{ $user }}{{ end }}
		{{- end }}
		`))
	if err != nil {
		return "", err
	}
	return tmpl.Render(template.Payload{
		"groups": auth.GetGroups(),
	})
}

type jobsCommand struct {
	noHandshake
	allowAll
	Help string
}

func (j jobsCommand) Execute(req request.Request) (string, error) {
	flags := flag.NewFlagSet("jobs", flag.ContinueOnError)
	limit := flags.Int("limit", 5, "how many jobs to return")
	if err := flags.Parse(req.Args); err != nil {
		return "", err
	}

	tmpl, err := template.New("jobs", strings.Join([]string{
		"{{- range $job := .jobs }}",
		"{{ HumanizeTime $job.StartTime }}",
		" - *{{ $job.Request.Command }}*",
		" by *{{ $job.Request.Username }}*",
		" in *{{ if $job.Request.IsIM }}DM{{ else }}{{ $job.Request.ChannelLink }}{{ end }}*",
		" - *{{ $job.Status }}*\n",
		"{{end}}",
	}, ""))
	if err != nil {
		return "", err
	}

	jobs, err := jobs.Latest(*limit)
	if err != nil {
		return "", err
	}
	return tmpl.Render(template.Payload{
		"jobs": jobs,
	})
}
