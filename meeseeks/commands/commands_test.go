package commands_test

import (
	"testing"

	"github.com/renstrom/dedent"
	"gitlab.com/mr-meeseeks/meeseeks-box/auth"
	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/commands"
	stubs "gitlab.com/mr-meeseeks/meeseeks-box/testingstubs"
	"gitlab.com/mr-meeseeks/meeseeks-box/version"
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

func Test_ShellCommand(t *testing.T) {
	cmds, err := commands.New(configWithEcho)
	stubs.Must(t, "shell command failed to build", err)

	cmd, err := cmds.Find("echo")
	stubs.Must(t, "find cmd failed", err)

	out, err := cmd.Execute("hello", "meeseeks")
	stubs.Must(t, "shell echo command erred out", err)
	stubs.AssertEquals(t, out, "hello meeseeks\n")
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

func Test_VersionCommand(t *testing.T) {
	cmds, err := commands.New(config.Config{})
	stubs.Must(t, "could not build commands", err)

	cmd, err := cmds.Find("version")
	stubs.Must(t, "failed to get version command", err)

	out, err := cmd.Execute()
	stubs.Must(t, "failed to execute version command", err)

	stubs.AssertEquals(t, version.AppVersion, out)
}

func Test_HelpCommand(t *testing.T) {
	cmds, err := commands.New(configWithEcho)
	stubs.Must(t, "could not build commands", err)

	cmd, err := cmds.Find("help")
	stubs.Must(t, "failed to get help command", err)

	out, err := cmd.Execute()
	stubs.Must(t, "failed to execute help command", err)

	stubs.AssertEquals(t, dedent.Dedent(`
		- echo: command that prints back the arguments passed
		- groups: prints the configured groups
		- help: prints all the kwnown commands and its associated help
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

	cmds, err := commands.New(configWithEcho)
	stubs.Must(t, "could not build commands", err)

	cmd, err := cmds.Find("groups")
	stubs.AssertEquals(t, cmd.HasHandshake(), false)
	stubs.AssertEquals(t, cmd.ConfiguredCommand().AuthStrategy, config.AuthStrategyAllowedGroup)
	stubs.Must(t, "failed to get help command", err)

	out, err := cmd.Execute()
	stubs.Must(t, "failed to execute help command", err)

	stubs.AssertEquals(t, dedent.Dedent(`
		- admins: admin_user
		- other: user_one, user_two
		`), out)
}
