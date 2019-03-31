package builtins

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitlab.com/yakshaving.art/meeseeks-box/auth"
	"gitlab.com/yakshaving.art/meeseeks-box/commands"
	"gitlab.com/yakshaving.art/meeseeks-box/meeseeks"
	"gitlab.com/yakshaving.art/meeseeks-box/persistence"
	"gitlab.com/yakshaving.art/meeseeks-box/text/template"
	"gitlab.com/yakshaving.art/meeseeks-box/version"
	"github.com/renstrom/dedent"
)

// Builtin Commands Names
const (
	BuiltinVersionCommand   = "version"
	BuiltinHelpCommand      = "help"
	BuiltinGroupsCommand    = "groups"
	BuiltinJobsCommand      = "jobs"
	BuiltinFindJobCommand   = "job"
	BuiltinAuditCommand     = "audit"
	BuiltinAuditJobCommand  = "auditjob"
	BuiltinAuditLogsCommand = "auditlogs"
	BuiltinLastCommand      = "last"
	BuiltinTailCommand      = "tail"
	BuiltinHeadCommand      = "head"
	BuiltinLogsCommand      = "logs"
	BuiltinCancelJobCommand = "cancel"
	BuiltinKillJobCommand   = "kill"

	BuiltinNewAPITokenCommand    = "token-new"
	BuiltinListAPITokenCommand   = "tokens"
	BuiltinRevokeAPITokenCommand = "token-revoke"

	BuiltinNewAliasCommand    = "alias"
	BuiltinDeleteAliasCommand = "unalias"
	BuiltinGetAliasesCommand  = "aliases"
)

// Commands is the basic set of builtin commands
var Commands = map[string]meeseeks.Command{
	// The help builtin command needs a pointer to the map of generated commands,
	// because of this it is added as the last one when building the whole command
	// map
	BuiltinVersionCommand: versionCommand{
		help: newHelp(
			"prints the running meeseeks version",
		),
		cmd: cmd{BuiltinVersionCommand},
	},
	BuiltinGroupsCommand: groupsCommand{
		help: newHelp(
			"prints the configured groups",
		),
		cmd: cmd{BuiltinGroupsCommand},
	},
	BuiltinJobsCommand: jobsCommand{
		help: newHelp(
			"shows the last executed jobs for the calling user",
			"-limit: how many jobs to show, 5 by default",
		),
		cmd: cmd{BuiltinJobsCommand},
	},
	BuiltinAuditCommand: auditCommand{
		help: newHelp(
			"lists jobs from all users or a specific one (admin only)",
			"-user: user to filter for",
			"-limit: how many jobs to show, 5 by default",
		),
		cmd: cmd{BuiltinAuditCommand},
	},
	BuiltinAuditJobCommand: auditJobCommand{
		help: newHelp(
			"shows a command metadata by job ID (admin only)",
			"-user: user to filter for",
			"job ID to look up for, mandatory",
		),
		cmd: cmd{BuiltinAuditJobCommand},
	},
	BuiltinAuditLogsCommand: auditLogsCommand{
		help: newHelp(
			"shows the logs of a job by ID (admin only)",
			"job ID to look up for, mandatory",
		),
		cmd: cmd{BuiltinAuditLogsCommand},
	},
	BuiltinLastCommand: lastCommand{
		help: newHelp(
			"shows the last job metadata executed by the current user",
		),
		cmd: cmd{BuiltinLastCommand},
	},
	BuiltinFindJobCommand: findJobCommand{
		help: newHelp(
			"show metadata of one job by id",
			"job ID to look for, mandatory",
		),
		cmd: cmd{BuiltinFindJobCommand},
	},
	BuiltinTailCommand: tailCommand{
		help: newHelp(
			"returns the last lines of the last executed job, or one selected by job ID",
			"-limit: how many lines to show",
			"job ID to look for, optional, if not provided the last executed one will be looked up",
		),
		cmd: cmd{BuiltinTailCommand},
	},
	BuiltinHeadCommand: headCommand{
		help: newHelp(
			"returns the top N log lines of a command output or error",
			"-limit: how many lines to show",
			"job ID to look for, optional, if not provided the last executed one will be looked up",
		),
		cmd: cmd{BuiltinHeadCommand},
	},
	BuiltinLogsCommand: logsCommand{
		help: newHelp(
			"returns the full output of the job passed as argument",
			"job ID to look for, mandatory",
		),
		cmd: cmd{BuiltinLogsCommand},
	},
	BuiltinNewAPITokenCommand: newAPITokenCommand{
		help: newHelp(
			"creates a new API token",
			"user that will be impersonated by the api, mandatory",
			"channel that will be used as the one in which the job was called",
			"command the token will be calling",
			"arguments to pass to the command",
		),
		cmd: cmd{BuiltinNewAPITokenCommand},
	},
	BuiltinListAPITokenCommand: listAPITokensCommand{
		help: newHelp(
			"lists the API tokens",
		),
		cmd: cmd{BuiltinListAPITokenCommand},
	},
	BuiltinRevokeAPITokenCommand: revokeAPITokenCommand{
		help: newHelp(
			"revokes an API token",
			"api token to revoke, mandatory",
		),
		cmd: cmd{BuiltinRevokeAPITokenCommand},
	},
	BuiltinNewAliasCommand: newAliasCommand{
		help: newHelp(
			"adds an alias for a command for the current user",
			"alias itself, mandatory",
			"command to alias, mandatory",
			"arguments to pass to the command when invoking the alias, optional",
		),
		cmd: cmd{BuiltinNewAliasCommand},
	},
	BuiltinDeleteAliasCommand: deleteAliasCommand{
		help: newHelp(
			"deletes an alias",
			"alias to delete, mandatory",
		),
		cmd: cmd{BuiltinDeleteAliasCommand},
	},
	BuiltinGetAliasesCommand: getAliasesCommand{
		help: newHelp(
			"list all the aliases for the current user",
		),
		cmd: cmd{BuiltinGetAliasesCommand},
	},
	BuiltinHelpCommand: helpCommand{
		help: newHelp(
			"shows the help for all the commands, or a single one",
			"-all: includes the builtin commands in the list",
			"command, optional, shows the extended help for a single command",
		),
		cmd: cmd{BuiltinHelpCommand},
	},

	// Added as a placeholder so they are recognized as a builtin command
	BuiltinCancelJobCommand: nil,
	BuiltinKillJobCommand:   nil,
}

var errNoJobIDAsArgument = fmt.Errorf("no job id passed")

// LoadBuiltins loads the builtin commands
func LoadBuiltins(cancelCommand, killCommand meeseeks.Command) error {
	Commands[BuiltinCancelJobCommand] = cancelCommand
	Commands[BuiltinKillJobCommand] = killCommand

	reg := make([]commands.CommandRegistration, 0)

	for name, cmd := range Commands {
		reg = append(reg, commands.CommandRegistration{
			Name: name,
			Cmd:  cmd,
		})
	}

	return commands.Register(
		commands.RegistrationArgs{
			Commands: reg,
			Kind:     commands.KindBuiltinCommand,
			Action:   commands.ActionRegister,
		})
}

type defaultTimeout struct{}

func (d defaultTimeout) GetTimeout() time.Duration {
	return meeseeks.DefaultCommandTimeout
}

type emptyArgs struct{}

func (b emptyArgs) GetArgs() []string {
	return []string{}
}

type noRecord struct{}

func (n noRecord) MustRecord() bool {
	return false
}

type imOnlyChannel struct{}

func (i imOnlyChannel) GetAllowedChannels() []string {
	return []string{}
}

func (i imOnlyChannel) GetChannelStrategy() string {
	return auth.ChannelStrategyIMOnly
}

type anyChannel struct{}

func (a anyChannel) GetAllowedChannels() []string {
	return []string{}
}

func (a anyChannel) GetChannelStrategy() string {
	return auth.ChannelStrategyAny
}

type allowAll struct{}

func (a allowAll) GetAuthStrategy() string {
	return auth.AuthStrategyAny
}

func (a allowAll) GetAllowedGroups() []string {
	return []string{}
}

type allowAdmins struct{}

func (a allowAdmins) GetAuthStrategy() string {
	return auth.AuthStrategyAllowedGroup
}

func (a allowAdmins) GetAllowedGroups() []string {
	return []string{auth.AdminGroup}
}

type noHandshake struct {
}

func (b noHandshake) HasHandshake() bool {
	return false
}

type cmd struct {
	cmd string
}

func (c cmd) GetCmd() string {
	return c.cmd
}

type versionCommand struct {
	cmd
	help
	noHandshake
	noRecord
	allowAll
	anyChannel
	emptyArgs
	defaultTimeout
}

func (v versionCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	return fmt.Sprintf("%s version %s, commit %s, built on %s",
		version.Name, version.Version, version.Commit, version.Date), nil
}

func newHelp(summary string, args ...string) help {
	return help{
		meeseeks.NewHelp(summary, args...),
	}
}

type help struct {
	commandHelp meeseeks.Help
}

func (h help) GetHelp() meeseeks.Help {
	return h.commandHelp
}

type helpCommand struct {
	cmd
	help
	noHandshake
	noRecord
	allowAll
	anyChannel
	emptyArgs
	defaultTimeout
}

var helpListTemplate = `{{ range $name, $c := .commands }}- {{ $name }}: {{ $c.GetHelp.GetSummary }}
{{ end }}`

var helpCommandTemplate = `*{{ .name }}* - {{ .help.GetSummary }}
{{ if gt ( len .help.GetArgs ) 0 }}
*Arguments*{{ range $a := .help.GetArgs }}
- {{ $a }}{{ end }}{{ end }}
`

func (h helpCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	flags := flag.NewFlagSet("help", flag.ContinueOnError)
	all := flags.Bool("all", false, "show help for all commands, including builtins")

	flags.Parse(job.Request.Args)

	switch flags.NArg() {
	case 0:
		tmpl, err := template.New("help", helpListTemplate)
		if err != nil {
			return "", err
		}

		cmds := make(map[string]meeseeks.Command)
		for k, c := range commands.All() {
			if _, isBuiltin := Commands[k]; isBuiltin && !*all {
				continue
			}
			cmds[k] = c
		}

		return tmpl.Render(map[string]interface{}{
			"commands": cmds,
		})

	case 1:
		if cmd, ok := commands.All()[flags.Arg(0)]; ok {
			tmpl, err := template.New("help", helpCommandTemplate)
			if err != nil {
				return "", err
			}
			return tmpl.Render(map[string]interface{}{
				"name": flags.Arg(0),
				"help": cmd.GetHelp(),
			})
		}
		return "", fmt.Errorf("could not find command %s", flags.Arg(0))

	default:
		return "", fmt.Errorf("too many arguments")
	}
}

type cancelJobCommand struct {
	cmd
	help
	noHandshake
	noRecord
	emptyArgs
	allowAll
	anyChannel
	defaultTimeout
	cancelFunc func(jobID uint64)
}

// NewCancelJobCommand creates a command that will invoke the passed cancel job function when executed
func NewCancelJobCommand(f func(jobID uint64)) meeseeks.Command {
	return cancelJobCommand{
		help: newHelp(
			"sends a cancellation signal to a job owned by the current user",
			"job ID to send the signal to",
		),
		cancelFunc: f,
	}
}

func (c cancelJobCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	jobID, err := parseJobID(job.Request.Args)
	if err != nil {
		return "", err
	}
	j, err := persistence.Jobs().Get(jobID)
	if err != nil {
		return "", err
	}
	if job.Request.Username != j.Request.Username {
		return "", meeseeks.ErrNoJobWithID
	}
	c.cancelFunc(jobID)
	return fmt.Sprintf("Issued command cancellation to job %d", jobID), nil
}

type killJobCommand struct {
	cmd
	help
	noHandshake
	noRecord
	emptyArgs
	allowAdmins
	anyChannel
	defaultTimeout
	cancelFunc func(jobID uint64)
}

// NewKillJobCommand creates a command that will invoke the passed cancel job function when executed
func NewKillJobCommand(f func(jobID uint64)) meeseeks.Command {
	return killJobCommand{
		help: newHelp(
			"sends a cancellation signal to a job, admin only",
			"job ID to send the signal to",
		),
		cancelFunc: f,
	}
}

func (k killJobCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	jobID, err := parseJobID(job.Request.Args)
	if err != nil {
		return "", err
	}
	_, err = persistence.Jobs().Get(jobID)
	if err != nil {
		return "", err
	}
	k.cancelFunc(jobID)
	return fmt.Sprintf("Issued command cancellation to job %d", jobID), nil
}

type groupsCommand struct {
	cmd
	help
	noHandshake
	noRecord
	emptyArgs
	allowAdmins
	anyChannel
	defaultTimeout
}

var groupsTemplate = dedent.Dedent(`
	{{- range $group, $users := .groups }}
	- {{ $group }}:
		{{- range $index, $user := $users }}{{ if ne $index 0 }},{{ end }} {{ $user }}{{ end }}
	{{- end }}
	`)

func (g groupsCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	tmpl, err := template.New("version", groupsTemplate)
	if err != nil {
		return "", err
	}
	return tmpl.Render(map[string]interface{}{
		"groups": auth.GetGroups(),
	})
}

type jobsCommand struct {
	cmd
	help
	noHandshake
	noRecord
	allowAll
	anyChannel
	emptyArgs
	defaultTimeout
}

var jobsTemplate = strings.Join([]string{
	"{{- $length := len .jobs }}{{- if eq $length 0 }}",
	"No jobs found\n",
	"{{ else }}",
	"{{- range $job := .jobs }}",
	"*{{ $job.ID }}* - {{ HumanizeTime $job.StartTime }}",
	" - *{{ $job.Request.Command }}*",
	" by *{{ $job.Request.Username }}*",
	" in *{{ if $job.Request.IsIM }}DM{{ else }}{{ $job.Request.ChannelLink }}{{ end }}*",
	" - *{{ $job.Status }}*\n",
	"{{ end }}",
	"{{ end }}",
}, "")

// jobMultiMatch builds a Match function from a list of Match functions
func jobsMultiMatch(matchers ...func(meeseeks.Job) bool) func(meeseeks.Job) bool {
	return func(job meeseeks.Job) bool {
		for _, matcher := range matchers {
			if !matcher(job) {
				return false
			}
		}
		return true
	}
}

func (j jobsCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	flags := flag.NewFlagSet("jobs", flag.ContinueOnError)
	limit := flags.Int("limit", 5, "how many jobs to return")
	status := flags.String("status", "", "filter jobs per status (running, failed or successful)")
	if err := flags.Parse(job.Request.Args); err != nil {
		return "", err
	}

	callingUser := job.Request.Username
	requestedStatus := strings.Title(*status)
	jobs, err := persistence.Jobs().Find(meeseeks.JobFilter{
		Limit: *limit,
		Match: jobsMultiMatch(
			isUser(callingUser),
			isStatusOrEmpty(requestedStatus),
		),
	})

	if err != nil {
		return "", err
	}
	tmpl, err := template.New("jobs", jobsTemplate)
	if err != nil {
		return "", err
	}
	return tmpl.Render(map[string]interface{}{
		"jobs": jobs,
	})
}

type auditCommand struct {
	cmd
	help
	noHandshake
	noRecord
	allowAdmins
	anyChannel
	emptyArgs
	defaultTimeout
}

func (j auditCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	flags := flag.NewFlagSet("audit", flag.ContinueOnError)
	limit := flags.Int("limit", 5, "how many jobs to return")
	user := flags.String("user", "", "the user to audit")
	status := flags.String("status", "", "filter jobs per status (running, failed or successful)")
	if err := flags.Parse(job.Request.Args); err != nil {
		return "", err
	}

	requestedStatus := strings.Title(*status)

	jobs, err := persistence.Jobs().Find(meeseeks.JobFilter{
		Limit: *limit,
		Match: jobsMultiMatch(
			isStatusOrEmpty(requestedStatus),
			func(j meeseeks.Job) bool {
				if *user == "" {
					return true
				}
				return *user == j.Request.Username
			},
		),
	})

	if err != nil {
		return "", err
	}
	tmpl, err := template.New("jobs", jobsTemplate)
	if err != nil {
		return "", err
	}
	return tmpl.Render(map[string]interface{}{
		"jobs": jobs,
	})
}

type lastCommand struct {
	cmd
	help
	noHandshake
	noRecord
	allowAll
	anyChannel
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

func (l lastCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	callingUser := job.Request.Username
	jobs, err := persistence.Jobs().Find(meeseeks.JobFilter{
		Limit: 1,
		Match: isUser(callingUser),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get the last job: %s", err)
	}
	if len(jobs) == 0 {
		return "", fmt.Errorf("no last command for current user")
	}
	tmpl, err := template.New("job", jobTemplate)
	if err != nil {
		return "", err
	}
	return tmpl.Render(map[string]interface{}{
		"job": jobs[0],
	})
}

type findJobCommand struct {
	cmd
	help
	noHandshake
	noRecord
	allowAll
	anyChannel
	emptyArgs
	defaultTimeout
}

func (l findJobCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	id, err := parseJobID(job.Request.Args)
	if err != nil {
		return "", err
	}

	callingUser := job.Request.Username
	jobs, err := persistence.Jobs().Find(meeseeks.JobFilter{
		Limit: 1,
		Match: jobsMultiMatch(
			isUser(callingUser),
			isJobID(id)),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get the last job: %s", err)
	}
	if len(jobs) == 0 {
		return "", fmt.Errorf("no last command for current user")
	}

	tmpl, err := template.New("job", jobTemplate)
	if err != nil {
		return "", err
	}
	return tmpl.Render(map[string]interface{}{
		"job": jobs[0],
	})
}

type auditJobCommand struct {
	cmd
	help
	noHandshake
	noRecord
	allowAdmins
	anyChannel
	emptyArgs
	defaultTimeout
}

func (l auditJobCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	id, err := parseJobID(job.Request.Args)
	if err != nil {
		return "", err
	}

	jobs, err := persistence.Jobs().Find(meeseeks.JobFilter{
		Limit: 1,
		Match: isJobID(id),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get job %d: %s", job.ID, err)
	}
	if len(jobs) == 0 {
		return "", fmt.Errorf("job not found")
	}

	tmpl, err := template.New("job", jobTemplate)
	if err != nil {
		return "", err
	}
	return tmpl.Render(map[string]interface{}{
		"job": jobs[0],
	})
}

type auditLogsCommand struct {
	cmd
	help
	noHandshake
	noRecord
	allowAll
	anyChannel
	emptyArgs
	defaultTimeout
}

func (t auditLogsCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	id, err := parseJobID(job.Request.Args)
	if err != nil {
		return "", err
	}

	jobs, err := persistence.Jobs().Find(meeseeks.JobFilter{
		Limit: 1,
		Match: isJobID(id),
	})

	if err != nil {
		return "", fmt.Errorf("failed to find job with id %d: %s", id, err)
	}
	if len(jobs) == 0 {
		return "", fmt.Errorf("there is no job %d", id)
	}
	j := jobs[0]

	jobLogs, err := persistence.LogReader().Get(j.ID)
	if err != nil {
		return "", err
	}
	return jobLogs.Output, jobLogs.GetError()
}

type tailCommand struct {
	cmd
	help
	noHandshake
	noRecord
	allowAll
	anyChannel
	emptyArgs
	defaultTimeout
}

func (t tailCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	flags := flag.NewFlagSet("tail", flag.ContinueOnError)
	limit := flags.Int("limit", 5, "how many lines to return")

	flags.Parse(job.Request.Args)

	jobID, err := parseJobID(flags.Args())

	if err == errNoJobIDAsArgument {
		jobID, err = findLastJobIDForUser(job.Request.Username)
	}
	if err != nil {
		return "", err
	}

	jobLogs, err := persistence.LogReader().Tail(jobID, *limit)
	if err != nil {
		return "", err
	}
	return jobLogs.Output, jobLogs.GetError()
}

type headCommand struct {
	cmd
	help
	noHandshake
	noRecord
	allowAll
	anyChannel
	emptyArgs
	defaultTimeout
}

func (h headCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	flags := flag.NewFlagSet("head", flag.ContinueOnError)
	limit := flags.Int("limit", 5, "how many lines to return")

	flags.Parse(job.Request.Args)

	jobID, err := parseJobID(flags.Args())

	if err == errNoJobIDAsArgument {
		jobID, err = findLastJobIDForUser(job.Request.Username)
	}
	if err != nil {
		return "", err
	}

	jobLogs, err := persistence.LogReader().Head(jobID, *limit)
	if err != nil {
		return "", err
	}
	return jobLogs.Output, jobLogs.GetError()
}

type logsCommand struct {
	cmd
	help
	noHandshake
	noRecord
	allowAll
	anyChannel
	emptyArgs
	defaultTimeout
}

func (t logsCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	id, err := parseJobID(job.Request.Args)
	if err != nil {
		return "", err
	}

	callingUser := job.Request.Username
	jobs, err := persistence.Jobs().Find(meeseeks.JobFilter{
		Limit: 1,
		Match: jobsMultiMatch(
			isUser(callingUser),
			isJobID(id)),
	})
	if err != nil {
		return "", fmt.Errorf("failed to find job with id %d: %s", id, err)
	}
	if len(jobs) == 0 {
		return "", fmt.Errorf("no job with id %d for user %s", id, callingUser)
	}
	j := jobs[0]

	jobLogs, err := persistence.LogReader().Get(j.ID)
	if err != nil {
		return "", err
	}
	return jobLogs.Output, jobLogs.GetError()
}

type newAPITokenCommand struct {
	cmd
	help
	noHandshake
	noRecord
	allowAdmins
	imOnlyChannel
	emptyArgs
	defaultTimeout
}

func (n newAPITokenCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	if len(job.Request.Args) < 3 {
		return "", fmt.Errorf("not enough arguments passed in")
	}

	t, err := persistence.APITokens().Create(
		job.Request.Args[0],
		job.Request.Args[1],
		strings.Join(job.Request.Args[2:], " "),
	)
	return fmt.Sprintf("created token %s", t), err
}

type revokeAPITokenCommand struct {
	cmd
	help
	noHandshake
	noRecord
	allowAdmins
	imOnlyChannel
	emptyArgs
	defaultTimeout
}

func (r revokeAPITokenCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	if len(job.Request.Args) != 1 {
		return "", fmt.Errorf("only one token ID should be passed as an argument")
	}
	tokenID := job.Request.Args[0]
	if err := persistence.APITokens().Revoke(tokenID); err != nil {
		return "", err
	}
	return fmt.Sprintf("Token *%s* has been revoked", tokenID), nil
}

type listAPITokensCommand struct {
	cmd
	help
	noHandshake
	noRecord
	allowAdmins
	imOnlyChannel
	emptyArgs
	defaultTimeout
}

var listTokensTemplate = `{{ if eq (len .tokens) 0 }}No tokens could be found{{ else }}{{ range $t := .tokens }}- *{{ $t.TokenID }}* {{ $t.UserLink }} at {{ $t.ChannelLink }} _{{ $t.Text}}_
{{ end }}{{ end }}`

// apiTokenMultiMatch builds a Match function from a list of Match functions
func apiTokenMultiMatch(matchers ...func(meeseeks.APIToken) bool) func(meeseeks.APIToken) bool {
	return func(token meeseeks.APIToken) bool {
		for _, matcher := range matchers {
			if !matcher(token) {
				return false
			}
		}
		return true
	}
}

func (l listAPITokensCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	flags := flag.NewFlagSet("jobs", flag.ContinueOnError)
	limit := flags.Int("limit", 5, "how many jobs to return")
	user := flags.String("user", "", "user to filter for")
	channel := flags.String("channel", "", "channel to filter for")
	command := flags.String("command", "", "command to filter for")

	flags.Parse(job.Request.Args)

	tmpl, err := template.New("tokens", listTokensTemplate)
	if err != nil {
		return "", err
	}

	matchers := []func(meeseeks.APIToken) bool{}
	if *user != "" {
		matchers = append(matchers, func(t meeseeks.APIToken) bool {
			return t.UserLink == *user
		})
	}
	if *channel != "" {
		matchers = append(matchers, func(t meeseeks.APIToken) bool {
			return t.ChannelLink == *channel
		})
	}
	if *command != "" {
		matchers = append(matchers, func(t meeseeks.APIToken) bool {
			return strings.HasPrefix(t.Text, *command)
		})
	}

	t, err := persistence.APITokens().Find(meeseeks.APITokenFilter{
		Limit: *limit,
		Match: apiTokenMultiMatch(matchers...),
	})
	if err != nil {
		return "", err
	}

	return tmpl.Render(map[string]interface{}{
		"tokens": t,
	})
}

type newAliasCommand struct {
	cmd
	help
	noHandshake
	noRecord
	allowAll
	anyChannel
	emptyArgs
	defaultTimeout
}

func (l newAliasCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	if len(job.Request.Args) < 2 {
		return "", fmt.Errorf("an alias requires at least two arguments: the alias and the command")
	}

	args := job.Request.Args
	if err := persistence.Aliases().Create(job.Request.UserID, args[0], args[1], args[2:]...); err != nil {
		return fmt.Sprintf("failed to create the alias. Error: %s", err), err
	}

	return "alias created successfully", nil
}

type deleteAliasCommand struct {
	cmd
	help
	noHandshake
	noRecord
	allowAll
	anyChannel
	emptyArgs
	defaultTimeout
}

func (l deleteAliasCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	if len(job.Request.Args) != 1 {
		return "", fmt.Errorf("unalias requires only one argument: the alias to delete")
	}

	if err := persistence.Aliases().Remove(job.Request.UserID, job.Request.Args[0]); err != nil {
		return fmt.Sprintf("failed to delete the alias. Error: %s", err), err
	}
	return "alias deleted successfully", nil

}

type getAliasesCommand struct {
	cmd
	help
	noHandshake
	noRecord
	allowAll
	anyChannel
	emptyArgs
	defaultTimeout
}

var getAliasesTemplate = `{{ if eq (len .aliases) 0 }}No alias could be found{{ else }}{{ range $a := .aliases }}- *{{ $a.Alias }}* - ` + "`" + `{{ $a.Command }}{{ range $arg := $a.Args }} {{ $arg }}{{ end }}` + "`" + `
{{ end }}{{ end }}`

func (l getAliasesCommand) Execute(_ context.Context, job meeseeks.Job) (string, error) {
	a, err := persistence.Aliases().List(job.Request.UserID)
	if err != nil {
		return fmt.Sprintf("failed to load the aliases. Error: %s", err), err
	}
	tmpl, err := template.New("aliases", getAliasesTemplate)
	if err != nil {
		return "", err
	}
	return tmpl.Render(map[string]interface{}{
		"aliases": a,
	})
}

// Helper functions from now on

func parseJobID(args []string) (uint64, error) {
	if len(args) == 0 {
		return 0, errNoJobIDAsArgument
	}
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid job ID %s: %s", args[0], err)
	}

	return id, nil
}

func isUser(username string) func(meeseeks.Job) bool {
	return func(j meeseeks.Job) bool {
		return j.Request.Username == username
	}
}

func isJobID(jobID uint64) func(meeseeks.Job) bool {
	return func(j meeseeks.Job) bool {
		return j.ID == jobID
	}
}

func isStatusOrEmpty(status string) func(meeseeks.Job) bool {
	return func(j meeseeks.Job) bool {
		if status == "" {
			return true
		}
		return j.Status == status
	}
}

func findLastJobIDForUser(callingUser string) (uint64, error) {
	jobs, err := persistence.Jobs().Find(meeseeks.JobFilter{
		Limit: 1,
		Match: isUser(callingUser),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get the last job: %s", err)
	}
	if len(jobs) == 0 {
		return 0, fmt.Errorf("no last command for current user")
	}
	return jobs[0].ID, nil

}
