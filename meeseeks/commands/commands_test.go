package commands_test

import (
	"testing"

	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/commands"
	stubs "gitlab.com/mr-meeseeks/meeseeks-box/testingstubs"
	"gitlab.com/mr-meeseeks/meeseeks-box/version"
)

func Test_ShellCommand(t *testing.T) {
	shell, err := commands.New(config.Command{
		Cmd:  "echo",
		Args: []string{},
		Type: config.ShellCommandType,
	})
	stubs.Must(t, "shell command failed to build", err)

	out, err := shell.Execute("hello", "meeseeks")
	stubs.Must(t, "shell echo command erred out", err)
	stubs.AssertEquals(t, out, "hello meeseeks\n")
}

func Test_InvalidCommand(t *testing.T) {
	_, err := commands.New(config.Command{
		Cmd:  "fail",
		Type: 0,
	})
	if err != commands.ErrCommandNotFound {
		t.Fatalf("command build should have failed with an error, got %s instead", err)
	}
}

func Test_VersionCommand(t *testing.T) {
	cmd, err := commands.New(config.BuiltinCommands[config.BuiltinCommandVersion])
	stubs.Must(t, "failed to get version command", err)

	out, err := cmd.Execute()
	stubs.Must(t, "failed to execute version command", err)

	stubs.AssertEquals(t, version.AppVersion, out)
}
