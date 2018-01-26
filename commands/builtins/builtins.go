package builtins

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pcarranza/meeseeks-box/jobs/logs"

	"github.com/pcarranza/meeseeks-box/jobs"

	"github.com/pcarranza/meeseeks-box/auth"
	"github.com/pcarranza/meeseeks-box/command"
	"github.com/pcarranza/meeseeks-box/template"
	"github.com/pcarranza/meeseeks-box/version"
	"github.com/renstrom/dedent"
)

// Builtin Commands Names
const (
	BuiltinVersionCommand  = "version"
	BuiltinHelpCommand     = "help"
	BuiltinGroupsCommand   = "groups"
	BuiltinJobsCommand     = "jobs"
	BuiltinFindJobCommand  = "job"
	BuiltinAuditCommand    = "audit"
	BuiltinAuditJobCommand = "auditjob"
	BuiltinLastCommand     = "last"
	BuiltinTailCommand     = "tail"
	BuiltinLogsCommand     = "logs"
)

// Commands is the basic set of builtin commands
var Commands = map[string]command.Command{
	// The help builtin command needs a pointer to the map of generated commands,
	// because of this it is added as the last one when building the whole command
	// map
	BuiltinVersionCommand: versionCommand{
		help: help{"prints the running meeseeks version"},
	},
	BuiltinGroupsCommand: groupsCommand{
		help: help{"prints the configured groups"},
	},
	BuiltinJobsCommand: jobsCommand{
		help: help{"shows the last executed jobs for the calling user"},
	},
	BuiltinAuditCommand: auditCommand{
		help: help{"find all jobs for all users or a specific one (admin only)"},
	},
	BuiltinLastCommand: lastCommand{
		help: help{"shows the last executed command by the calling user"},
	},
	BuiltinFindJobCommand: findJob{
		help: help{"find one job"},
	},
	BuiltinAuditJobCommand: auditJobCommand{
		help: help{"shows a specific command by the specified user (admin only)"},
	},
	BuiltinTailCommand: tailCommand{
		help: help{"returns the last command output or error"},
	},
	BuiltinLogsCommand: logsCommand{
		help: help{"returns the logs of the command id passed as argument"},
	},
}

// AddHelpCommand creates a new help command and adds it to the map
func AddHelpCommand(c map[string]command.Command) {
	c[BuiltinHelpCommand] = helpCommand{
		commands: c,
		help:     help{"prints all the kwnown commands and its associated help"},
	}
}

type plainTemplates struct{}

func (p plainTemplates) Templates() map[string]string {
	return map[string]string{
		template.SuccessKey: fmt.Sprintf("{{ .user }} {{ AnyValue \"%s\" . }}{{ with $out := .output }}\n{{ $out }}{{ end }}", template.SuccessKey),
	}
}

type defaultTemplates struct {
}

func (d defaultTemplates) Templates() map[string]string {
	return template.GetDefaultTemplates()
}

type defaultTimeout struct{}

func (d defaultTimeout) Timeout() time.Duration {
	return command.DefaultCommandTimeout
}

type emptyArgs struct{}

func (b emptyArgs) Args() []string {
	return []string{}
}

type noRecord struct{}

func (n noRecord) Record() bool {
	return false
}

type allowAll struct{}

func (a allowAll) AuthStrategy() string {
	return auth.AuthStrategyAny
}

func (a allowAll) AllowedGroups() []string {
	return []string{}
}

type allowAdmins struct{}

func (a allowAdmins) AuthStrategy() string {
	return auth.AuthStrategyAllowedGroup
}

func (a allowAdmins) AllowedGroups() []string {
	return []string{auth.AdminGroup}
}

type noHandshake struct {
}

func (b noHandshake) HasHandshake() bool {
	return false
}

type help struct {
	help string
}

func (h help) Help() string {
	return h.help
}

type versionCommand struct {
	help
	noHandshake
	noRecord
	allowAll
	plainTemplates
	emptyArgs
	defaultTimeout
}

func (v versionCommand) Cmd() string {
	return BuiltinVersionCommand
}

func (v versionCommand) Execute(job jobs.Job) (string, error) {
	return fmt.Sprintf("meeseeks-box version %s, commit %s, built at %s",
		version.Version, version.Commit, version.Date), nil
}

type helpCommand struct {
	help
	noHandshake
	noRecord
	allowAll
	plainTemplates
	emptyArgs
	defaultTimeout
	commands map[string]command.Command
}

var helpTemplate = dedent.Dedent(`
	{{ range $name, $cmd := .commands }}- {{ $name }}: {{ $cmd.Help }}
	{{ end }}`)

func (h helpCommand) Cmd() string {
	return BuiltinHelpCommand
}

func (h helpCommand) Execute(job jobs.Job) (string, error) {
	tmpl, err := template.New("help", helpTemplate)
	if err != nil {
		return "", err
	}
	return tmpl.Render(template.Payload{
		"commands": h.commands,
	})
}

type groupsCommand struct {
	help
	noHandshake
	noRecord
	emptyArgs
	allowAdmins
	plainTemplates
	defaultTimeout
}

var groupsTemplate = dedent.Dedent(`
	{{- range $group, $users := .groups }}
	- {{ $group }}:
		{{- range $index, $user := $users }}{{ if ne $index 0 }},{{ end }} {{ $user }}{{ end }}
	{{- end }}
	`)

func (g groupsCommand) Cmd() string {
	return BuiltinGroupsCommand
}

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
	help
	noHandshake
	noRecord
	allowAll
	plainTemplates
	emptyArgs
	defaultTimeout
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

func (j jobsCommand) Cmd() string {
	return BuiltinJobsCommand
}

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

type auditCommand struct {
	help
	noHandshake
	noRecord
	allowAdmins
	plainTemplates
	emptyArgs
	defaultTimeout
}

func (j auditCommand) Cmd() string {
	return BuiltinAuditCommand
}

func (j auditCommand) Execute(job jobs.Job) (string, error) {
	flags := flag.NewFlagSet("audit", flag.ContinueOnError)
	limit := flags.Int("limit", 5, "how many jobs to return")
	user := flags.String("user", "", "the user to audit")
	if err := flags.Parse(job.Request.Args); err != nil {
		return "", err
	}

	jobs, err := jobs.Find(jobs.JobFilter{
		Limit: *limit,
		Match: func(j jobs.Job) bool {
			if *user == "" {
				return true
			}
			return *user == j.Request.Username
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
	help
	noHandshake
	noRecord
	allowAll
	plainTemplates
	emptyArgs
	defaultTimeout
}

var jobTemplate = `
{{- with $job := .job }}{{ with $r := $job.Request }}* *ID* {{ $job.ID }}
* *Status* {{ $job.Status}}
* *Command* {{ $r.Command }}{{ with $args := $r.Args }}
* *Args* "{{ Join $args "\" \"" }}" {{ end }}
* *Where* {{ if $r.IsIM }}IM{{ else }}{{ $r.ChannelLink }}{{ end }}
* *When* {{ HumanizeTime $job.StartTime }}
{{- end }}{{- end }}
`

func (l lastCommand) Cmd() string {
	return BuiltinLastCommand
}

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
	help
	noHandshake
	noRecord
	allowAll
	plainTemplates
	emptyArgs
	defaultTimeout
}

func (l findJob) Cmd() string {
	return BuiltinFindJobCommand
}

func (l findJob) Execute(job jobs.Job) (string, error) {
	id, err := parseJobID(job)
	if err != nil {
		return "", err
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

type auditJobCommand struct {
	help
	noHandshake
	noRecord
	allowAdmins
	plainTemplates
	emptyArgs
	defaultTimeout
}

func (l auditJobCommand) Cmd() string {
	return BuiltinAuditJobCommand
}

func (l auditJobCommand) Execute(job jobs.Job) (string, error) {
	flags := flag.NewFlagSet("auditjobs", flag.ContinueOnError)
	if err := flags.Parse(job.Request.Args); err != nil {
		return "", err
	}
	id, err := parseJobID(job)
	if err != nil {
		return "", err
	}

	jobs, err := jobs.Find(jobs.JobFilter{
		Limit: 1,
		Match: func(j jobs.Job) bool {
			return j.ID == id
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get job %d: %s", job.ID, err)
	}
	if len(jobs) == 0 {
		return "", fmt.Errorf("Job not found")
	}

	tmpl, err := template.New("job", jobTemplate)
	if err != nil {
		return "", err
	}
	return tmpl.Render(template.Payload{
		"job": jobs[0],
	})
}

type tailCommand struct {
	help
	noHandshake
	noRecord
	allowAll
	defaultTemplates
	emptyArgs
	defaultTimeout
}

func (t tailCommand) Cmd() string {
	return BuiltinTailCommand
}

func (t tailCommand) Execute(job jobs.Job) (string, error) {
	callingUser := job.Request.Username
	jobs, err := jobs.Find(jobs.JobFilter{
		Limit: 1,
		Match: func(j jobs.Job) bool {
			return j.Request.Username == callingUser &&
				j.Request.Command != BuiltinTailCommand
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get the last job: %s", err)
	}
	if len(jobs) == 0 {
		return "", fmt.Errorf("No last command for current user")
	}
	j := jobs[0]

	jobLogs, err := logs.Get(j.ID)
	if err != nil {
		return "", err
	}
	return jobLogs.Output, jobLogs.GetError()
}

type logsCommand struct {
	help
	noHandshake
	noRecord
	allowAll
	defaultTemplates
	emptyArgs
	defaultTimeout
}

func (t logsCommand) Cmd() string {
	return BuiltinTailCommand
}

func (t logsCommand) Execute(job jobs.Job) (string, error) {
	id, err := parseJobID(job)
	if err != nil {
		return "", err
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
		return "", fmt.Errorf("failed to find job with id %d: %s", id, err)
	}
	if len(jobs) == 0 {
		return "", fmt.Errorf("No job with id %d for user %s", id, callingUser)
	}
	j := jobs[0]

	jobLogs, err := logs.Get(j.ID)
	if err != nil {
		return "", err
	}
	return jobLogs.Output, jobLogs.GetError()
}

func parseJobID(job jobs.Job) (uint64, error) {
	if len(job.Request.Args) == 0 {
		return 0, fmt.Errorf("no job id passed")
	}
	id, err := strconv.ParseUint(job.Request.Args[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid job ID %s: %s", job.Request.Args[0], err)
	}

	return id, nil
}
