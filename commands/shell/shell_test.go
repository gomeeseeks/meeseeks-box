package shell_test

import (
	"testing"

	"github.com/pcarranza/meeseeks-box/command"
	"github.com/pcarranza/meeseeks-box/commands/shell"
	"github.com/pcarranza/meeseeks-box/jobs"
	"github.com/pcarranza/meeseeks-box/meeseeks/request"
	stubs "github.com/pcarranza/meeseeks-box/testingstubs"
)

var echoCommand = shell.New(shell.CommandOpts{
	Cmd:  "echo",
	Help: "command that prints back the arguments passed",
})

var failCommand = shell.New(shell.CommandOpts{
	Cmd:  "false",
	Help: "command that fails",
})

func TestShellCommand(t *testing.T) {
	stubs.AssertEquals(t, "echo", echoCommand.Cmd())
	stubs.AssertEquals(t, []string{}, echoCommand.Args())
	stubs.AssertEquals(t, []string{}, echoCommand.AllowedGroups())
	stubs.AssertEquals(t, true, echoCommand.HasHandshake())
	stubs.AssertEquals(t, true, echoCommand.Record())
	stubs.AssertEquals(t, map[string]string{}, echoCommand.Templates())
	stubs.AssertEquals(t, command.DefaultCommandTimeout, echoCommand.Timeout())
	stubs.AssertEquals(t, "command that prints back the arguments passed", echoCommand.Help())
}

func TestExecuteEcho(t *testing.T) {
	stubs.WithTmpDB(func() {
		out, err := echoCommand.Execute(jobs.Job{
			ID:      1,
			Request: request.Request{Args: []string{"hello", "meeseeks\nsecond line"}},
		})
		stubs.Must(t, "failed to execute echo command", err)
		stubs.AssertEquals(t, "hello meeseeks\nsecond line\n", out)
	})
}

func TestExecuteFail(t *testing.T) {
	stubs.WithTmpDB(func() {
		_, err := failCommand.Execute(jobs.Job{
			ID:      2,
			Request: request.Request{},
		})
		stubs.AssertEquals(t, "exit status 1", err.Error())
	})
}
