package builtins_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gomeeseeks/meeseeks-box/auth"
	"github.com/gomeeseeks/meeseeks-box/commands"
	"github.com/gomeeseeks/meeseeks-box/commands/builtins"
	"github.com/gomeeseeks/meeseeks-box/commands/shell"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/mocks"
	"github.com/gomeeseeks/meeseeks-box/persistence"
)

var basicGroups = map[string][]string{
	"admins": {"admin_user"},
	"other":  {"user_one", "user_two"},
}

var req = meeseeks.Request{
	Command:     "command",
	Channel:     "general",
	ChannelID:   "123",
	ChannelLink: "<#123>",
	UserID:      "someoneID",
	Username:    "someone",
	Args:        []string{"arg1", "arg2"},
}

func Test_BuiltinCommands(t *testing.T) {
	auth.Configure(basicGroups)
	commands.LoadBuiltins()

	commands.Replace(commands.CommandRegistration{
		Name: builtins.BuiltinCancelJobCommand, Cmd: builtins.NewCancelJobCommand(
			func(_ uint64) {})})
	commands.Replace(commands.CommandRegistration{
		Name: builtins.BuiltinKillJobCommand, Cmd: builtins.NewKillJobCommand(
			func(_ uint64) {})})

	tt := []struct {
		name                    string
		req                     meeseeks.Request
		job                     meeseeks.Job
		setup                   func()
		expected                string
		expectedMatch           string
		expectedError           error
		expectedAuthStrategy    string
		expectedAllowedGroups   []string
		expectedChannelStrategy string
		expectedAllowedChannels []string
	}{
		{
			name: "version command",
			req: meeseeks.Request{
				Command: builtins.BuiltinVersionCommand,
				UserID:  "userid",
			},

			job:                     meeseeks.Job{},
			expected:                "meeseeks-box version , commit , built on ",
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "help non builtins command",
			req: meeseeks.Request{
				Command: builtins.BuiltinHelpCommand,
				UserID:  "userid",
			},

			job:                     meeseeks.Job{Request: meeseeks.Request{Args: []string{}}},
			expected:                "",
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "help all command",
			req: meeseeks.Request{
				Command: builtins.BuiltinHelpCommand,
				UserID:  "userid",
			},
			job: meeseeks.Job{Request: meeseeks.Request{Args: []string{"-all"}}},
			expected: `- alias: adds an alias for a command for the current user
- aliases: list all the aliases for the current user
- audit: lists jobs from all users or a specific one (admin only)
- auditjob: shows a command metadata by job ID (admin only)
- auditlogs: shows the logs of a job by ID (admin only)
- cancel: sends a cancellation signal to a job owned by the current user
- groups: prints the configured groups
- head: returns the top N log lines of a command output or error
- help: shows the help for all the commands, or a single one
- job: show metadata of one job by id
- jobs: shows the last executed jobs for the calling user
- kill: sends a cancellation signal to a job, admin only
- last: shows the last job metadata executed by the current user
- logs: returns the full output of the job passed as argument
- tail: returns the last lines of the last executed job, or one selected by job ID
- token-new: creates a new API token
- token-revoke: revokes an API token
- tokens: lists the API tokens
- unalias: deletes an alias
- version: prints the running meeseeks version
`,
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "help one command",
			req: meeseeks.Request{
				Command: builtins.BuiltinHelpCommand,
				UserID:  "userid",
			},
			job: meeseeks.Job{Request: meeseeks.Request{Args: []string{"token-new"}}},
			expected: `*token-new* - creates a new API token

*Arguments*
- user that will be impersonated by the api, mandatory
- channel that will be used as the one in which the job was called
- command the token will be calling
- arguments to pass to the command
`,
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "groups command",
			req: meeseeks.Request{
				Command: builtins.BuiltinGroupsCommand,
				UserID:  "userid",
			},

			job: meeseeks.Job{},
			expected: `
- admins: admin_user
- other: user_one, user_two
`,
			expectedAuthStrategy:    auth.AuthStrategyAllowedGroup,
			expectedAllowedGroups:   []string{auth.AdminGroup},
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test jobs command",
			req: meeseeks.Request{
				Command: builtins.BuiltinJobsCommand,
				UserID:  "userid",
			},

			job: meeseeks.Job{
				Request: meeseeks.Request{Username: "someone"},
			},
			setup: func() {
				j, err := persistence.Jobs().Create(req)
				mocks.Must(t, "could not create job", err)
				persistence.Jobs().Succeed(j.ID)
			},
			expected:                "*1* - now - *command* by *someone* in *<#123>* - *Successful*\n",
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test audit command",
			req: meeseeks.Request{
				Command: builtins.BuiltinAuditCommand,
				UserID:  "userid",
			},

			job: meeseeks.Job{
				Request: meeseeks.Request{},
			},
			setup: func() {
				j, err := persistence.Jobs().Create(req)
				mocks.Must(t, "could not create job", err)
				persistence.Jobs().Succeed(j.ID)
			},
			expected:                "*1* - now - *command* by *someone* in *<#123>* - *Successful*\n",
			expectedAuthStrategy:    auth.AuthStrategyAllowedGroup,
			expectedAllowedGroups:   []string{auth.AdminGroup},
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test jobs command with limit",
			req: meeseeks.Request{
				Command: builtins.BuiltinJobsCommand,
				UserID:  "userid",
			},

			job: meeseeks.Job{
				Request: meeseeks.Request{Username: "someone", Args: []string{"-limit=1"}},
			},
			setup: func() {
				persistence.Jobs().Create(req)
				persistence.Jobs().Create(req)
			},
			expected:                "*2* - now - *command* by *someone* in *<#123>* - *Running*\n",
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test jobs command on IM",
			req: meeseeks.Request{
				Command: builtins.BuiltinJobsCommand,
				UserID:  "userid",
			},

			job: meeseeks.Job{
				Request: meeseeks.Request{Username: "someone"},
			},
			setup: func() {
				persistence.Jobs().Create(meeseeks.Request{
					Command:   "command",
					Channel:   "general",
					ChannelID: "123",
					Username:  "someone",
					Args:      []string{"arg1", "arg2"},
					IsIM:      true,
				})
			},
			expected:                "*1* - now - *command* by *someone* in *DM* - *Running*\n",
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test last command",
			req: meeseeks.Request{
				Command: builtins.BuiltinLastCommand,
				UserID:  "userid",
			},

			job: meeseeks.Job{
				Request: meeseeks.Request{Username: "someone"},
			},
			setup: func() {
				persistence.Jobs().Create(req)
				persistence.Jobs().Create(req)
				persistence.Jobs().Create(req)
			},
			expected:                "* *ID* 3\n* *Status* Running\n* *Command* command\n* *Args* \"arg1\" \"arg2\" \n* *Where* <#123>\n* *When* now\n",
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test find command",
			req: meeseeks.Request{
				Command: builtins.BuiltinFindJobCommand,
				UserID:  "userid",
			},

			job: meeseeks.Job{
				Request: meeseeks.Request{Username: "someone", Args: []string{"1"}},
			},
			setup: func() {
				persistence.Jobs().Create(req)
				persistence.Jobs().Create(req)
			},
			expected:                "* *ID* 1\n* *Status* Running\n* *Command* command\n* *Args* \"arg1\" \"arg2\" \n* *Where* <#123>\n* *When* now\n",
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test alias command",
			req: meeseeks.Request{
				Command: builtins.BuiltinNewAliasCommand,
				UserID:  "userid",
			},

			job: meeseeks.Job{
				Request: meeseeks.Request{
					Command: "alias",
					Args:    []string{"command", "for", "-add", "alias"},
					UserID:  "userid",
				}},
			expected:                "alias created successfully",
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test aliases command",
			req: meeseeks.Request{
				Command: builtins.BuiltinGetAliasesCommand,
				UserID:  "userid",
			},

			job: meeseeks.Job{
				Request: meeseeks.Request{
					Command: "aliases",
					UserID:  "userid",
				},
			},
			setup: func() {
				err := persistence.Aliases().Create("userid", "first", "command", []string{"-with args"}...)
				mocks.Must(t, "create first alias", err)

				err = persistence.Aliases().Create("userid", "second", "another", []string{"-command"}...)
				mocks.Must(t, "create second alias", err)
			},
			expected:                "- *first* - `command -with args`\n- *second* - `another -command`\n",
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test unalias command",
			req: meeseeks.Request{
				Command: builtins.BuiltinGetAliasesCommand,
				UserID:  "userid",
			},

			job: meeseeks.Job{
				Request: meeseeks.Request{
					Command: "aliases",
					UserID:  "userid",
				},
			},
			setup: func() {
				err := persistence.Aliases().Create("userid", "command", "command", []string{"-with", "args"}...)
				mocks.Must(t, "create first alias", err)

				err = persistence.Aliases().Create("userid", "second", "another", []string{"-command"}...)
				mocks.Must(t, "create second alias", err)

				err = persistence.Aliases().Remove("userid", "command")
				mocks.Must(t, "delete first alias", err)

			},
			expected:                "- *second* - `another -command`\n",
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test alias execution",
			req: meeseeks.Request{
				Command: "testalias",
				UserID:  "userid",
			},

			job: meeseeks.Job{
				Request: meeseeks.Request{
					Command: "testalias",
					UserID:  "userid",
				},
			},
			setup: func() {
				err := persistence.Aliases().Create("userid", "testalias", "audit", []string{"-limit", "1"}...)
				mocks.Must(t, "create an alias", err)

				commands.Add(
					commands.CommandRegistration{
						Name: "noop", Cmd: shell.New(meeseeks.CommandOpts{
							AuthStrategy: "any",
							Cmd:          "true",
						})})

				_, err = persistence.Jobs().Create(
					meeseeks.Request{
						Command: "noop",
						UserID:  "userid",
					})
				mocks.Must(t, "do nothing", err)

			},
			expected:                "*1* - now - *noop* by ** in ** - *Running*\n",
			expectedAuthStrategy:    auth.AuthStrategyAllowedGroup,
			expectedAllowedGroups:   []string{auth.AdminGroup},
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test auditjob command",
			req: meeseeks.Request{
				Command: builtins.BuiltinAuditJobCommand,
				UserID:  "userid",
			},

			job: meeseeks.Job{
				Request: meeseeks.Request{Username: "someone", Args: []string{"1"}},
			},
			setup: func() {
				persistence.Jobs().Create(req)
				persistence.Jobs().Create(req)
			},
			expected:                "* *ID* 1\n* *Status* Running\n* *Command* command\n* *Args* \"arg1\" \"arg2\" \n* *Where* <#123>\n* *When* now\n",
			expectedAuthStrategy:    auth.AuthStrategyAllowedGroup,
			expectedAllowedGroups:   []string{auth.AdminGroup},
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test tail command with jobID",
			req: meeseeks.Request{
				Command: builtins.BuiltinTailCommand,
				UserID:  "someone",
			},
			job: meeseeks.Job{
				Request: meeseeks.Request{
					Username: "someone",
					Args:     []string{"-limit", "1", "1"}},
			},
			setup: func() {

				j, err := persistence.Jobs().Create(req)
				mocks.Must(t, "create job", err)

				w := persistence.LogWriter()
				w.Append(j.ID, "line 1.1")
				w.Append(j.ID, "line 1.2")
				w.Append(j.ID, "line 1.3")
				w.Append(j.ID, "line 1.4")

				j, err = persistence.Jobs().Create(req)
				mocks.Must(t, "create job", err)

				w = persistence.LogWriter()
				w.Append(j.ID, "line 2.1")
				w.Append(j.ID, "line 2.2")
				w.Append(j.ID, "line 2.3")
			},
			expected:                "line 1.4",
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test tail command",
			req: meeseeks.Request{
				Command: builtins.BuiltinTailCommand,
				UserID:  "someone",
			},

			job: meeseeks.Job{
				Request: meeseeks.Request{Username: "someone", Args: []string{"-limit", "2"}},
			},
			setup: func() {

				j, err := persistence.Jobs().Create(req)
				mocks.Must(t, "create job", err)

				w := persistence.LogWriter()
				w.Append(j.ID, "line 1.1")
				w.Append(j.ID, "line 1.2")
				w.Append(j.ID, "line 1.3")

				j, err = persistence.Jobs().Create(req)
				mocks.Must(t, "create job", err)

				w = persistence.LogWriter()
				w.Append(j.ID, "line 2.1")
				w.Append(j.ID, "line 2.2")
				w.Append(j.ID, "line 2.3")
			},
			expected:                "line 2.2\nline 2.3",
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test head command",
			req: meeseeks.Request{
				Command: builtins.BuiltinHeadCommand,
				UserID:  "userid",
			},
			job: meeseeks.Job{
				Request: meeseeks.Request{Username: "someone", Args: []string{"-limit", "2"}},
			},
			setup: func() {
				j, err := persistence.Jobs().Create(req)
				mocks.Must(t, "create job", err)

				w := persistence.LogWriter()
				w.Append(j.ID, "line 1.1\nline 1.2\nsomething to say 1")

				j, err = persistence.Jobs().Create(req)
				mocks.Must(t, "create job", err)

				w = persistence.LogWriter()
				w.Append(j.ID, "line 2.1")
				w.Append(j.ID, "line 2.2")
				w.Append(j.ID, "something to say 2")
			},
			expected:                "line 2.1\nline 2.2",
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test head command with jobID",
			req: meeseeks.Request{
				Command: builtins.BuiltinHeadCommand,
				UserID:  "userid",
			},
			job: meeseeks.Job{
				Request: meeseeks.Request{Username: "someone", Args: []string{"-limit", "1", "1"}},
			},
			setup: func() {
				j, err := persistence.Jobs().Create(req)
				mocks.Must(t, "create job", err)

				w := persistence.LogWriter()
				w.Append(j.ID, "line 1.1")
				w.Append(j.ID, "line 1.2")
				w.Append(j.ID, "something to say 1")

				j, err = persistence.Jobs().Create(req)
				mocks.Must(t, "create job", err)

				w = persistence.LogWriter()
				w.Append(j.ID, "line 2.1")
				w.Append(j.ID, "line 2.2")
				w.Append(j.ID, "something to say 2")
			},
			expected:                "line 1.1",
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test logs command",
			req: meeseeks.Request{
				Command: builtins.BuiltinLogsCommand,
				UserID:  "userid",
			},

			job: meeseeks.Job{
				Request: meeseeks.Request{Username: "someone", Args: []string{"1"}},
			},
			setup: func() {
				j, err := persistence.Jobs().Create(req)
				mocks.Must(t, "create job", err)
				w := persistence.LogWriter()
				w.Append(j.ID, "something to say 1")

				j, err = persistence.Jobs().Create(req)
				mocks.Must(t, "create job", err)
				w = persistence.LogWriter()
				w.Append(j.ID, "something to say 2")
			},
			expected:                "something to say 1",
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test auditlogs command",
			req: meeseeks.Request{
				Command: builtins.BuiltinAuditLogsCommand,
				UserID:  "userid",
			},

			job: meeseeks.Job{
				Request: meeseeks.Request{Username: "admin_user", Args: []string{"1"}},
			},
			setup: func() {
				j, err := persistence.Jobs().Create(req)
				mocks.Must(t, "create job", err)
				w := persistence.LogWriter()
				w.Append(j.ID, "something to say 1")

				j, err = persistence.Jobs().Create(req)
				mocks.Must(t, "create job", err)
				w = persistence.LogWriter()
				w.Append(j.ID, "something to say 2")
			},
			expected:                "something to say 1",
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test token-new command",
			req: meeseeks.Request{
				Command: builtins.BuiltinNewAPITokenCommand,
				UserID:  "userid",
			},

			job: meeseeks.Job{
				Request: meeseeks.Request{Username: "admin_user", IsIM: true, Args: []string{"admin_user", "yolo", "rm", "-rf"}},
			},
			expectedMatch:           "created token .*",
			expectedAuthStrategy:    auth.AuthStrategyAllowedGroup,
			expectedAllowedGroups:   []string{auth.AdminGroup},
			expectedChannelStrategy: auth.ChannelStrategyIMOnly,
		},
		{
			name: "test tokens command",
			req: meeseeks.Request{
				Command: builtins.BuiltinListAPITokenCommand,
				UserID:  "userid",
			},

			job: meeseeks.Job{
				Request: meeseeks.Request{Username: "admin_user", IsIM: true},
			},
			setup: func() {
				_, err := persistence.APITokens().Create(
					"userLink",
					"channelLink",
					"something",
				)
				mocks.Must(t, "create token", err)

			},
			expectedMatch:           "- \\*.*?\\* userLink at channelLink _something_",
			expectedAuthStrategy:    auth.AuthStrategyAllowedGroup,
			expectedAllowedGroups:   []string{auth.AdminGroup},
			expectedChannelStrategy: auth.ChannelStrategyIMOnly,
		},
		{
			name: "test kill job command",
			req: meeseeks.Request{
				Command: builtins.BuiltinKillJobCommand,
				UserID:  "userid",
			},

			job: meeseeks.Job{
				Request: meeseeks.Request{
					Username: "someone",
					Args:     []string{"1"}},
			},
			setup: func() {
				_, err := persistence.Jobs().Create(req)
				mocks.Must(t, "create job", err)

				_, err = persistence.Jobs().Create(req)
				mocks.Must(t, "create job", err)
			},
			expected:                "Issued command cancellation to job 1",
			expectedAuthStrategy:    auth.AuthStrategyAllowedGroup,
			expectedAllowedGroups:   []string{auth.AdminGroup},
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test cancel job command",
			req: meeseeks.Request{
				Command: builtins.BuiltinCancelJobCommand,
				UserID:  "userid",
			},

			job: meeseeks.Job{
				Request: meeseeks.Request{Username: "someone", Args: []string{"2"}},
			},
			setup: func() {
				_, err := persistence.Jobs().Create(req)
				mocks.Must(t, "create job", err)

				_, err = persistence.Jobs().Create(req)
				mocks.Must(t, "create job", err)
			},
			expected:                "Issued command cancellation to job 2",
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
		{
			name: "test cancel job command with wrong user",
			req: meeseeks.Request{
				Command: builtins.BuiltinCancelJobCommand,
				UserID:  "userid",
			},
			job: meeseeks.Job{
				Request: meeseeks.Request{Username: "someone_else",
					Args: []string{"2"},
				},
			},
			setup: func() {
				_, err := persistence.Jobs().Create(req)
				mocks.Must(t, "create job", err)

				_, err = persistence.Jobs().Create(req)
				mocks.Must(t, "create job", err)
			},
			expectedError:           fmt.Errorf("no job could be found"),
			expectedAuthStrategy:    auth.AuthStrategyAny,
			expectedChannelStrategy: auth.ChannelStrategyAny,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			mocks.Must(t, "failed to run tests", mocks.WithTmpDB(func(_ string) {
				if tc.setup != nil {
					tc.setup()
				}
				cmd, ok := commands.Find(&tc.req)
				if !ok {
					t.Fatalf("could not find command %s", tc.req.Command)
				}

				// mocks.AssertEquals(t, cmd.GetCmd(), tc.req.Command)
				mocks.AssertEquals(t, cmd.GetTimeout(), meeseeks.DefaultCommandTimeout)
				mocks.AssertEquals(t, []string{}, cmd.GetAllowedChannels())
				mocks.AssertEquals(t, []string{}, cmd.GetArgs())
				mocks.AssertEquals(t, false, cmd.MustRecord())
				mocks.AssertEquals(t, false, cmd.HasHandshake())
				mocks.AssertEquals(t, tc.expectedAuthStrategy, cmd.GetAuthStrategy())
				switch tc.expectedAuthStrategy {
				case auth.AuthStrategyAllowedGroup:
					mocks.AssertEquals(t, tc.expectedAllowedGroups, cmd.GetAllowedGroups())
				default:
					mocks.AssertEquals(t, []string{}, cmd.GetAllowedGroups())
				}

				mocks.AssertEquals(t, tc.expectedChannelStrategy, cmd.GetChannelStrategy())
				switch tc.expectedChannelStrategy {
				case auth.ChannelStrategyAllowedChannels:
					mocks.AssertEquals(t, tc.expectedAllowedChannels, cmd.GetAllowedChannels())
				default:
					mocks.AssertEquals(t, []string{}, cmd.GetAllowedChannels())

				}

				out, err := cmd.Execute(context.Background(), tc.job)
				if err != nil && tc.expectedError != nil {
					mocks.AssertEquals(t, err.Error(), tc.expectedError.Error())
					return
				}

				mocks.Must(t, "cmd erred out", err)
				if tc.expected != "" {
					mocks.AssertEquals(t, tc.expected, out)
				}
				if tc.expectedMatch != "" {
					mocks.AssertMatches(t, tc.expectedMatch, out)
				}
			}))
		})
	}
}

func Test_FilterJobsAudit(t *testing.T) {
	mocks.Must(t, "failed to audit the correct jobs", mocks.WithTmpDB(func(_ string) {
		r1 := meeseeks.Request{
			Command:     "command",
			Channel:     "general",
			ChannelID:   "123",
			ChannelLink: "<#123>",
			Username:    "someone",
			Args:        []string{"some", "thing"},
		}
		r2 := meeseeks.Request{
			Command:     "command",
			Channel:     "general",
			ChannelID:   "123",
			ChannelLink: "<#123>",
			Username:    "someoneelse",
			Args:        []string{"something", "else"},
		}

		auth.Configure(basicGroups)
		persistence.Jobs().Create(r1)
		persistence.Jobs().Create(r2)
		persistence.Jobs().Create(r1)
		persistence.Jobs().Create(r1)
		persistence.Jobs().Create(r2)

		cmd, ok := commands.Find(&meeseeks.Request{
			Command: "audit",
			UserID:  "userid",
		})
		if !ok {
			t.Fatalf("could not find command %s", "audit")
		}

		audit, err := cmd.Execute(context.Background(), meeseeks.Job{
			Request: meeseeks.Request{Args: []string{"-user", "someone"}},
		})
		if err != nil {
			t.Fatalf("Failed to execute audit: %s", err)
		}
		mocks.AssertEquals(t, "*4* - now - *command* by *someone* in *<#123>* - *Running*\n*3* - now - *command* by *someone* in *<#123>* - *Running*\n*1* - now - *command* by *someone* in *<#123>* - *Running*\n", audit)

		limit, err := cmd.Execute(context.Background(), meeseeks.Job{
			Request: meeseeks.Request{Args: []string{"-user", "someone", "-limit", "2"}},
		})
		if err != nil {
			t.Fatalf("Failed to execute audit: %s", err)
		}
		mocks.AssertEquals(t, "*4* - now - *command* by *someone* in *<#123>* - *Running*\n*3* - now - *command* by *someone* in *<#123>* - *Running*\n", limit)
	}))
}

func TestAPITokenLifecycle(t *testing.T) {
	exec := func(r meeseeks.Request) (string, error) {
		cmd, ok := commands.Find(&r)
		if !ok {
			t.Fatalf("could not find command %s", r.Command)
		}

		return cmd.Execute(context.Background(), persistence.Jobs().Null(r))
	}

	mocks.Must(t, "failed to audit the correct jobs", mocks.WithTmpDB(func(_ string) {

		out, err := exec(meeseeks.Request{
			Command: builtins.BuiltinListAPITokenCommand,
			UserID:  "apiuser",
			IsIM:    true,
		})
		mocks.Must(t, "can't list api tokens:", err)
		mocks.AssertEquals(t, "No tokens could be found", out)

		out, err = exec(meeseeks.Request{
			Command: builtins.BuiltinNewAPITokenCommand,
			UserID:  "apiuser",
			IsIM:    true,
			Args:    []string{"apiuser", "yolo", "rm", "-rf"},
		})
		mocks.Must(t, "can't create an api token:", err)

		token := strings.Split(out, " ")[2]

		out, err = exec(meeseeks.Request{
			Command: builtins.BuiltinListAPITokenCommand,
			UserID:  "apiuser",
			IsIM:    true,
		})
		mocks.Must(t, "can't list api tokens:", err)
		mocks.AssertEquals(t, fmt.Sprintf("- *%s* apiuser at yolo _rm -rf_\n", token), out)

		out, err = exec(meeseeks.Request{
			Command: builtins.BuiltinRevokeAPITokenCommand,
			UserID:  "apiuser",
			IsIM:    true,
			Args:    []string{token},
		})
		mocks.Must(t, "can't revoke an api token:", err)
		mocks.AssertEquals(t, fmt.Sprintf("Token *%s* has been revoked", token), out)

		out, err = exec(meeseeks.Request{
			Command: builtins.BuiltinListAPITokenCommand,
			UserID:  "apiuser",
			IsIM:    true,
		})
		mocks.Must(t, "can't list api tokens:", err)
		mocks.AssertEquals(t, "No tokens could be found", out)
	}))
}
