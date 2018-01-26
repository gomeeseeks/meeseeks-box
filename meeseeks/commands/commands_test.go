package commands_test

import (
	"testing"

	"github.com/pcarranza/meeseeks-box/jobs/logs"

	"github.com/pcarranza/meeseeks-box/auth"
	"github.com/pcarranza/meeseeks-box/config"
	"github.com/pcarranza/meeseeks-box/jobs"
	"github.com/renstrom/dedent"

	cmds "github.com/pcarranza/meeseeks-box/commands"
	"github.com/pcarranza/meeseeks-box/meeseeks/commands"
	"github.com/pcarranza/meeseeks-box/meeseeks/request"

	stubs "github.com/pcarranza/meeseeks-box/testingstubs"
)

var configWithEcho = config.Config{
	Commands: map[string]config.Command{
		"echo": config.Command{
			Cmd:     "echo",
			Args:    []string{},
			Timeout: cmds.DefaultCommandTimeout,
			Type:    config.ShellCommandType,
			Help:    "command that prints back the arguments passed",
		},
	},
	Groups: map[string][]string{
		"admins": []string{"admin_user"},
		"other":  []string{"user_one", "user_two"},
	},
}

var req = request.Request{
	Command:     "command",
	Channel:     "general",
	ChannelID:   "123",
	ChannelLink: "<#123>",
	Username:    "someone",
	Args:        []string{"arg1", "arg2"},
}

func Test_Commands(t *testing.T) {
	cmds, err := commands.New(configWithEcho)
	stubs.Must(t, "failed to create commands", err)

	auth.Configure(configWithEcho.Groups)

	tt := []struct {
		name     string
		cmd      string
		job      jobs.Job
		setup    func()
		expected string
	}{
		{
			name: "shell command",
			cmd:  "echo",
			job: jobs.Job{
				Request: request.Request{Args: []string{"hello", "meeseeks\nsecond line"}},
			},
			expected: "hello meeseeks\nsecond line\n",
		},
		{
			name:     "version command",
			cmd:      commands.BuiltinVersionCommand,
			job:      jobs.Job{},
			expected: "meeseeks-box version , commit , built at ",
		},
		{
			name: "help command",
			cmd:  commands.BuiltinHelpCommand,
			job:  jobs.Job{},
			expected: dedent.Dedent(`
				- audit: find all jobs for all users or a specific one (admin only)
				- auditjob: shows a specific command by the specified user (admin only)
				- echo: command that prints back the arguments passed
				- groups: prints the configured groups
				- help: prints all the kwnown commands and its associated help
				- job: find one job
				- jobs: shows the last executed jobs for the calling user
				- last: shows the last executed command by the calling user
				- logs: returns the logs of the command id passed as argument
				- tail: returns the last command output or error
				- version: prints the running meeseeks version
				`),
		},
		{
			name: "groups command",
			cmd:  commands.BuiltinGroupsCommand,
			job:  jobs.Job{},
			expected: dedent.Dedent(`
					- admins: admin_user
					- other: user_one, user_two
					`),
		},
		{
			name: "test jobs command",
			cmd:  commands.BuiltinJobsCommand,
			job: jobs.Job{
				Request: request.Request{Username: "someone"},
			},
			setup: func() {
				j, err := jobs.Create(req)
				stubs.Must(t, "could not create job", err)
				j.Finish(jobs.SuccessStatus)
			},
			expected: "*1* - now - *command* by *someone* in *<#123>* - *Successful*\n",
		},
		{
			name: "test audit command",
			cmd:  commands.BuiltinAuditCommand,
			job: jobs.Job{
				Request: request.Request{},
			},
			setup: func() {
				j, err := jobs.Create(req)
				stubs.Must(t, "could not create job", err)
				j.Finish(jobs.SuccessStatus)
			},
			expected: "*1* - now - *command* by *someone* in *<#123>* - *Successful*\n",
		},
		{
			name: "test jobs command with limit",
			cmd:  commands.BuiltinJobsCommand,
			job: jobs.Job{
				Request: request.Request{Username: "someone", Args: []string{"-limit=1"}},
			},
			setup: func() {
				jobs.Create(req)
				jobs.Create(req)
			},
			expected: "*2* - now - *command* by *someone* in *<#123>* - *Running*\n",
		},
		{
			name: "test jobs command on IM",
			cmd:  commands.BuiltinJobsCommand,
			job: jobs.Job{
				Request: request.Request{Username: "someone"},
			},
			setup: func() {
				jobs.Create(request.Request{
					Command:   "command",
					Channel:   "general",
					ChannelID: "123",
					Username:  "someone",
					Args:      []string{"arg1", "arg2"},
					IsIM:      true,
				})
			},
			expected: "*1* - now - *command* by *someone* in *DM* - *Running*\n",
		},
		{
			name: "test last command",
			cmd:  commands.BuiltinLastCommand,
			job: jobs.Job{
				Request: request.Request{Username: "someone"},
			},
			setup: func() {
				jobs.Create(req)
				jobs.Create(req)
				jobs.Create(req)
			},
			expected: "* *ID* 3\n* *Status* Running\n* *Command* command\n* *Args* \"arg1\" \"arg2\" \n* *Where* <#123>\n* *When* now\n",
		},
		{
			name: "test find command",
			cmd:  commands.BuiltinFindJobCommand,
			job: jobs.Job{
				Request: request.Request{Username: "someone", Args: []string{"1"}},
			},
			setup: func() {
				jobs.Create(req)
				jobs.Create(req)
			},
			expected: "* *ID* 1\n* *Status* Running\n* *Command* command\n* *Args* \"arg1\" \"arg2\" \n* *Where* <#123>\n* *When* now\n",
		},
		{
			name: "test auditjob command",
			cmd:  commands.BuiltinAuditJobCommand,
			job: jobs.Job{
				Request: request.Request{Username: "someone", Args: []string{"1"}},
			},
			setup: func() {
				jobs.Create(req)
				jobs.Create(req)
			},
			expected: "* *ID* 1\n* *Status* Running\n* *Command* command\n* *Args* \"arg1\" \"arg2\" \n* *Where* <#123>\n* *When* now\n",
		},
		{
			name: "test tail command",
			cmd:  commands.BuiltinTailCommand,
			job: jobs.Job{
				Request: request.Request{Username: "someone"},
			},
			setup: func() {
				j, err := jobs.Create(req)
				stubs.Must(t, "create job", err)
				logs.Append(j.ID, "something to say 1")

				j, err = jobs.Create(req)
				stubs.Must(t, "create job", err)
				logs.Append(j.ID, "something to say 2")
			},
			expected: "something to say 2",
		},
		{
			name: "test logs command",
			cmd:  commands.BuiltinLogsCommand,
			job: jobs.Job{
				Request: request.Request{Username: "someone", Args: []string{"1"}},
			},
			setup: func() {
				j, err := jobs.Create(req)
				stubs.Must(t, "create job", err)
				logs.Append(j.ID, "something to say 1")

				j, err = jobs.Create(req)
				stubs.Must(t, "create job", err)
				logs.Append(j.ID, "something to say 2")
			},
			expected: "something to say 1",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			stubs.Must(t, "failed to run tests", stubs.WithTmpDB(func() {
				if tc.setup != nil {
					tc.setup()
				}
				cmd, err := cmds.Find(tc.cmd)
				stubs.Must(t, "cmd failed", err)

				out, err := cmd.Execute(tc.job)
				stubs.Must(t, "cmd erred out", err)
				stubs.AssertEquals(t, tc.expected, out)
			}))
		})
	}
}

func Test_InvalidCommand(t *testing.T) {
	cmds, err := commands.New(
		config.Config{
			Commands: map[string]config.Command{},
		})
	stubs.Must(t, "could not build commands", err)
	_, err = cmds.Find("non-existing")
	if err != commands.ErrCommandNotFound {
		t.Fatalf("command build should have failed with an error, got %s instead", err)
	}
}

func Test_FilterJobsAudit(t *testing.T) {
	stubs.Must(t, "failed to audit the correct jobs", stubs.WithTmpDB(func() {
		r1 := request.Request{
			Command:     "command",
			Channel:     "general",
			ChannelID:   "123",
			ChannelLink: "<#123>",
			Username:    "someone",
			Args:        []string{"some", "thing"},
		}
		r2 := request.Request{
			Command:     "command",
			Channel:     "general",
			ChannelID:   "123",
			ChannelLink: "<#123>",
			Username:    "someoneelse",
			Args:        []string{"something", "else"},
		}

		cmds, err := commands.New(configWithEcho)
		stubs.Must(t, "failed to create commands", err)

		auth.Configure(configWithEcho.Groups)
		jobs.Create(r1)
		jobs.Create(r2)
		jobs.Create(r1)
		jobs.Create(r1)
		jobs.Create(r2)

		cmd, err := cmds.Find("audit")
		stubs.Must(t, "cmd failed", err)

		audit, err := cmd.Execute(jobs.Job{
			Request: request.Request{Args: []string{"-user", "someone"}},
		})
		if err != nil {
			t.Fatalf("Failed to execute audit: %s", err)
		}
		stubs.AssertEquals(t, "*4* - now - *command* by *someone* in *<#123>* - *Running*\n*3* - now - *command* by *someone* in *<#123>* - *Running*\n*1* - now - *command* by *someone* in *<#123>* - *Running*\n", audit)

		limit, err := cmd.Execute(jobs.Job{
			Request: request.Request{Args: []string{"-user", "someone", "-limit", "2"}},
		})
		if err != nil {
			t.Fatalf("Failed to execute audit: %s", err)
		}
		stubs.AssertEquals(t, "*4* - now - *command* by *someone* in *<#123>* - *Running*\n*3* - now - *command* by *someone* in *<#123>* - *Running*\n", limit)
	}))
}
