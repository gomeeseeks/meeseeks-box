package commands_test

import (
	"strings"
	"testing"

	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/commands"
	stubs "gitlab.com/mr-meeseeks/meeseeks-box/testingstubs"
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
	if err == nil || !strings.HasPrefix(err.Error(), "could not build command from") {
		t.Fatalf("command build should have failed with an error, got %s instead", err)
	}
}
