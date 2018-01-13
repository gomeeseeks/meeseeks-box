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

// Builtin Commands Names
const (
	BuiltinVersionCommand = "version"
	BuiltinHelpCommand    = "help"
	BuiltinGroupsCommand  = "groups"
	BuiltinJobsCommand    = "jobs"
	BuiltinLastCommand    = "last"
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

var builtInCommands = map[string]Command{
	// The help builtin command needs a pointer to the map of generated commands,
	// because of this it is added as the last one when building the whole command
	// map
	BuiltinVersionCommand: versionCommand{
		Help: "prints the running meeseeks version",
	},
	BuiltinGroupsCommand: groupsCommand{
		Help: "prints the configured groups",
	},
	BuiltinJobsCommand: jobsCommand{
		Help: "shows the last executed jobs",
	},
	BuiltinLastCommand: lastCommand{
		Help: "shows the last executed command by the invoking user",
	},
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

var helpTemplate = dedent.Dedent(
	`{{ range $name, $cmd := .commands }}
	- {{ $name }}: {{ $cmd.Help }}
	{{- end }}
	`)

func (h helpCommand) Execute(req request.Request) (string, error) {
	tmpl, err := template.New("version", helpTemplate)
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

var groupsTemplate = dedent.Dedent(`
	{{- range $group, $users := .groups }}
	- {{ $group }}:
		{{- range $index, $user := $users }}{{ if ne $index 0 }},{{ end }} {{ $user }}{{ end }}
	{{- end }}
	`)

func (g groupsCommand) Execute(req request.Request) (string, error) {
	tmpl, err := template.New("version", groupsTemplate)
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

var jobsTemplate = strings.Join([]string{
	"{{- range $job := .jobs }}",
	"{{ HumanizeTime $job.StartTime }}",
	" - *{{ $job.Request.Command }}*",
	" by *{{ $job.Request.Username }}*",
	" in *{{ if $job.Request.IsIM }}DM{{ else }}{{ $job.Request.ChannelLink }}{{ end }}*",
	" - *{{ $job.Status }}*\n",
	"{{end}}",
}, "")

func (j jobsCommand) Execute(req request.Request) (string, error) {
	flags := flag.NewFlagSet("jobs", flag.ContinueOnError)
	limit := flags.Int("limit", 5, "how many jobs to return")
	if err := flags.Parse(req.Args); err != nil {
		return "", err
	}

	tmpl, err := template.New("jobs", jobsTemplate)
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

type lastCommand struct {
	noHandshake
	allowAll
	Help string
}

var commandTemplate = `{{with $job := .job}}{{ with $r := $job.Request }}* *Command* {{ $r.Command }}
* *Args* {{ Join $r.Args ", " }}
* *Status* {{ $job.Status}}
* *Where* {{ if $r.IsIM }}IM{{ else }}{{ $r.ChannelLink }}{{end}}
* *When* {{ HumanizeTime $job.StartTime }}
* *ID* {{ $job.ID }}
{{- end }}{{- end }}
`

func (l lastCommand) Execute(req request.Request) (string, error) {
	tmpl, err := template.New("job", commandTemplate)
	if err != nil {
		return "", err
	}

	j, err := jobs.Last(req.Username, req.Command)
	if err != nil {
		return "", fmt.Errorf("failed to get the last job: %s", err)
	}
	return tmpl.Render(template.Payload{
		"job": j,
	})
}
