package command_test

import (
	"testing"

	"github.com/renstrom/dedent"
	"gitlab.com/mr-meeseeks/meeseeks-box/auth"
	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	"gitlab.com/mr-meeseeks/meeseeks-box/jobs"
	"gitlab.com/mr-meeseeks/meeseeks-box/version"

	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/command"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/request"

	stubs "gitlab.com/mr-meeseeks/meeseeks-box/testingstubs"
)

var configWithEcho = config.Config{
	Commands: map[string]config.Command{
		"echo": config.Command{
			Cmd:     "echo",
			Args:    []string{},
			Timeout: config.DefaultCommandTimeout,
			Type:    config.ShellCommandType,
			Help:    "command that prints back the arguments passed",
		},
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

func Test_ShellCommand(t *testing.T) {
	cmds, err := command.New(configWithEcho)
	stubs.Must(t, "shell command failed to build", err)

	cmd, err := cmds.Find("echo")
	stubs.Must(t, "find cmd failed", err)

	out, err := cmd.Execute(request.Request{Args: []string{"hello", "meeseeks"}})
	stubs.Must(t, "shell echo command erred out", err)
	stubs.AssertEquals(t, out, "hello meeseeks\n")
}

func Test_InvalidCommand(t *testing.T) {
	cmds, err := command.New(
		config.Config{
			Commands: map[string]config.Command{},
		})
	stubs.Must(t, "could not build commands", err)
	_, err = cmds.Find("non-existing")
	if err != command.ErrCommandNotFound {
		t.Fatalf("command build should have failed with an error, got %s instead", err)
	}
}

func Test_VersionCommand(t *testing.T) {
	cmds, err := command.New(config.Config{})
	stubs.Must(t, "could not build commands", err)

	cmd, err := cmds.Find("version")
	stubs.Must(t, "failed to get version command", err)

	out, err := cmd.Execute(request.Request{})
	stubs.Must(t, "failed to execute version command", err)

	stubs.AssertEquals(t, version.Version, out)
}

func Test_HelpCommand(t *testing.T) {
	cmds, err := command.New(configWithEcho)
	stubs.Must(t, "could not build commands", err)

	cmd, err := cmds.Find("help")
	stubs.Must(t, "failed to get help command", err)

	out, err := cmd.Execute(request.Request{})
	stubs.Must(t, "failed to execute help command", err)

	stubs.AssertEquals(t, dedent.Dedent(`
		- echo: command that prints back the arguments passed
		- groups: prints the configured groups
		- help: prints all the kwnown commands and its associated help
		- job: find one job
		- jobs: shows the last executed jobs for the calling user
		- last: shows the last executed command by the calling user
		- version: prints the running meeseeks version
		`), out)
}

func Test_GroupsCommand(t *testing.T) {
	auth.Configure(config.Config{
		Groups: map[string][]string{
			"admins": []string{"admin_user"},
			"other":  []string{"user_one", "user_two"},
		},
	})

	cmds, err := command.New(configWithEcho)
	stubs.Must(t, "could not build commands", err)

	cmd, err := cmds.Find("groups")
	stubs.Must(t, "failed to get help command", err)
	stubs.AssertEquals(t, cmd.HasHandshake(), false)
	stubs.AssertEquals(t, cmd.ConfiguredCommand().AuthStrategy, config.AuthStrategyAllowedGroup)

	out, err := cmd.Execute(request.Request{})
	stubs.Must(t, "failed to execute help command", err)

	stubs.AssertEquals(t, dedent.Dedent(`
		- admins: admin_user
		- other: user_one, user_two
		`), out)
}

func Test_JobsCommand(t *testing.T) {
	stubs.Must(t, "failed to run tests", stubs.WithTmpDB(func() {
		j, err := jobs.Create(req)
		stubs.Must(t, "could not create job", err)
		jobs.Finish(j.ID, jobs.SuccessStatus)

		cmds, err := command.New(configWithEcho)
		stubs.Must(t, "could not build commands", err)

		cmd, err := cmds.Find("jobs")
		stubs.Must(t, "failed to get jobs command", err)
		stubs.AssertEquals(t, cmd.HasHandshake(), false)
		stubs.AssertEquals(t, cmd.ConfiguredCommand().AuthStrategy, config.AuthStrategyAny)

		out, err := cmd.Execute(request.Request{Username: "someone"})
		stubs.Must(t, "failed to execute help command", err)

		stubs.AssertEquals(t, "*1* - now - *command* by *someone* in *<#123>* - *Successful*\n", out)
	}))
}

func Test_JobsCommandWithIM(t *testing.T) {
	stubs.Must(t, "failed to run tests", stubs.WithTmpDB(func() {
		jobs.Create(request.Request{
			Command:   "command",
			Channel:   "general",
			ChannelID: "123",
			Username:  "someone",
			Args:      []string{"arg1", "arg2"},
			IsIM:      true,
		})
		cmds, err := command.New(configWithEcho)
		stubs.Must(t, "could not build commands", err)

		cmd, err := cmds.Find("jobs")
		stubs.Must(t, "failed to get jobs command", err)
		stubs.AssertEquals(t, cmd.HasHandshake(), false)
		stubs.AssertEquals(t, cmd.ConfiguredCommand().AuthStrategy, config.AuthStrategyAny)

		out, err := cmd.Execute(request.Request{Username: "someone"})
		stubs.Must(t, "failed to execute help command", err)

		stubs.AssertEquals(t, "*1* - now - *command* by *someone* in *DM* - *Running*\n", out)
	}))
}
func Test_JobsChangeLimit(t *testing.T) {
	stubs.Must(t, "failed to run tests", stubs.WithTmpDB(func() {
		jobs.Create(req)
		jobs.Create(req)

		cmds, err := command.New(configWithEcho)
		stubs.Must(t, "could not build commands", err)

		cmd, err := cmds.Find("jobs")
		stubs.Must(t, "failed to get jobs command", err)
		stubs.AssertEquals(t, cmd.HasHandshake(), false)
		stubs.AssertEquals(t, cmd.ConfiguredCommand().AuthStrategy, config.AuthStrategyAny)

		out, err := cmd.Execute(request.Request{Username: "someone"})
		stubs.Must(t, "failed to execute help command", err)

		stubs.AssertEquals(t, "*2* - now - *command* by *someone* in *<#123>* - *Running*\n*1* - now - *command* by *someone* in *<#123>* - *Running*\n", out)

		out, err = cmd.Execute(request.Request{Username: "someone", Args: []string{"-limit=1"}})
		stubs.Must(t, "failed to execute help command", err)

		stubs.AssertEquals(t, "*2* - now - *command* by *someone* in *<#123>* - *Running*\n", out)
	}))
}

func Test_LastCommand(t *testing.T) {
	stubs.Must(t, "failed to run tests", stubs.WithTmpDB(func() {
		_, err := jobs.Create(req)
		stubs.Must(t, "could not create job", err)

		cmds, err := command.New(configWithEcho)
		stubs.Must(t, "could not build commands", err)

		cmd, err := cmds.Find(command.BuiltinLastCommand)
		stubs.Must(t, "failed to get jobs command", err)
		stubs.AssertEquals(t, cmd.HasHandshake(), false)
		stubs.AssertEquals(t, cmd.ConfiguredCommand().AuthStrategy, config.AuthStrategyAny)

		out, err := cmd.Execute(request.Request{
			Username: req.Username,
			Command:  command.BuiltinLastCommand,
		})
		stubs.Must(t, "failed to execute last command", err)

		stubs.AssertEquals(t, "* *Command* command\n* *Args* arg1, arg2\n* *Status* Running\n* *Where* <#123>\n* *When* now\n* *ID* 1\n", out)
	}))
}

func Test_FindJobCommand(t *testing.T) {
	stubs.Must(t, "failed to run tests", stubs.WithTmpDB(func() {
		j, err := jobs.Create(req)
		stubs.Must(t, "could not create job", err)
		jobs.Finish(j.ID, jobs.SuccessStatus)

		cmds, err := command.New(configWithEcho)
		stubs.Must(t, "could not build commands", err)

		cmd, err := cmds.Find(command.BuiltinFindJobCommand)
		stubs.Must(t, "failed to get jobs command", err)
		stubs.AssertEquals(t, cmd.HasHandshake(), false)
		stubs.AssertEquals(t, cmd.ConfiguredCommand().AuthStrategy, config.AuthStrategyAny)

		out, err := cmd.Execute(request.Request{Username: "someone", Args: []string{"1"}})
		stubs.Must(t, "failed to execute help command", err)

		stubs.AssertEquals(t, "* *Command* command\n* *Args* arg1, arg2\n* *Status* Successful\n* *Where* <#123>\n* *When* now\n* *ID* 1\n", out)
	}))
}
