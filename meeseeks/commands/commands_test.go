package commands_test

import (
	"testing"

	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/commands"
	stubs "gitlab.com/mr-meeseeks/meeseeks-box/testingstubs"
	"gitlab.com/mr-meeseeks/meeseeks-box/version"
)

func Test_ShellCommand(t *testing.T) {
	cmds, err := commands.New(
		config.Config{
			Commands: map[string]config.Command{
				"echo": config.Command{
					Cmd:     "echo",
					Args:    []string{},
					Timeout: config.DefaultCommandTimeout,
					Type:    config.ShellCommandType,
				},
			},
		})
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
	cmds, err := commands.New(config.Config{})
	stubs.Must(t, "could not build commands", err)

	cmd, err := cmds.Find("help")
	stubs.Must(t, "failed to get help command", err)

	out, err := cmd.Execute()
	stubs.Must(t, "failed to execute help command", err)

	stubs.AssertEquals(t, "invalid", out)
}
