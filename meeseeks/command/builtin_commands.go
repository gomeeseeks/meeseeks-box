package command

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"gitlab.com/mr-meeseeks/meeseeks-box/jobs"

	"github.com/renstrom/dedent"
	"gitlab.com/mr-meeseeks/meeseeks-box/auth"
	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/template"
	"gitlab.com/mr-meeseeks/meeseeks-box/version"
)

// Builtin Commands Names
const (
	BuiltinVersionCommand = "version"
	BuiltinHelpCommand    = "help"
	BuiltinGroupsCommand  = "groups"
	BuiltinJobsCommand    = "jobs"
	BuiltinFindJobCommand = "job"
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
		Help: "shows the last executed jobs for the calling user",
	},
	BuiltinLastCommand: lastCommand{
		Help: "shows the last executed command by the calling user",
	},
	BuiltinFindJobCommand: findJob{
		Help: "find one job",
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

func (v versionCommand) Execute(job jobs.Job) (string, error) {
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

func (h helpCommand) Execute(job jobs.Job) (string, error) {
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

func (g groupsCommand) Execute(job jobs.Job) (string, error) {
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
	"*{{ $job.ID }}* - {{ HumanizeTime $job.StartTime }}",
	" - *{{ $job.Request.Command }}*",
	" by *{{ $job.Request.Username }}*",
	" in *{{ if $job.Request.IsIM }}DM{{ else }}{{ $job.Request.ChannelLink }}{{ end }}*",
	" - *{{ $job.Status }}*\n",
	"{{end}}",
}, "")

func (j jobsCommand) Execute(job jobs.Job) (string, error) {
	flags := flag.NewFlagSet("jobs", flag.ContinueOnError)
	limit := flags.Int("limit", 5, "how many jobs to return")
	if err := flags.Parse(job.Request.Args); err != nil {
		return "", err
	}

	callingUser := job.Request.Username
	jobs, err := jobs.Find(jobs.JobFilter{
		Limit: *limit,
		Match: func(j jobs.Job) bool {
			return callingUser == j.Request.Username
		},
	})

	if err != nil {
		return "", err
	}
	tmpl, err := template.New("jobs", jobsTemplate)
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

var jobTemplate = `
{{- with $job := .job }}{{ with $r := $job.Request }}* *Command* {{ $r.Command }}{{ with $args := $r.Args }}
* *Args* "{{ Join $args "\" \"" }}" {{ end }}
* *Status* {{ $job.Status}}
* *Where* {{ if $r.IsIM }}IM{{ else }}{{ $r.ChannelLink }}{{ end }}
* *When* {{ HumanizeTime $job.StartTime }}
* *ID* {{ $job.ID }}
{{- end }}{{- end }}
`

func (l lastCommand) Execute(job jobs.Job) (string, error) {
	callingUser := job.Request.Username
	jobs, err := jobs.Find(jobs.JobFilter{
		Limit: 1,
		Match: func(j jobs.Job) bool {
			return j.Request.Username == callingUser &&
				j.Request.Command != BuiltinLastCommand
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get the last job: %s", err)
	}
	if len(jobs) == 0 {
		return "", fmt.Errorf("No last command for current user")
	}
	tmpl, err := template.New("job", jobTemplate)
	if err != nil {
		return "", err
	}
	return tmpl.Render(template.Payload{
		"job": jobs[0],
	})
}

type findJob struct {
	noHandshake
	allowAll
	Help string
}

func (l findJob) Execute(job jobs.Job) (string, error) {
	if len(job.Request.Args) == 0 {
		return "", fmt.Errorf("No job id to search for")
	}
	id, err := strconv.ParseUint(job.Request.Args[0], 10, 64)
	if err != nil {
		return "", fmt.Errorf("Invalid job ID %s: %s", job.Request.Args[0], err)
	}

	callingUser := job.Request.Username
	jobs, err := jobs.Find(jobs.JobFilter{
		Limit: 1,
		Match: func(j jobs.Job) bool {
			return j.Request.Username == callingUser &&
				j.ID == id
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get the last job: %s", err)
	}
	if len(jobs) == 0 {
		return "", fmt.Errorf("No last command for current user")
	}

	tmpl, err := template.New("job", jobTemplate)
	if err != nil {
		return "", err
	}
	return tmpl.Render(template.Payload{
		"job": jobs[0],
	})
}
