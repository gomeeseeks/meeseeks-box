package shell_test

import (
	"testing"

	"github.com/pcarranza/meeseeks-box/command"
	"github.com/pcarranza/meeseeks-box/commands/shell"
	"github.com/pcarranza/meeseeks-box/jobs"
	"github.com/pcarranza/meeseeks-box/meeseeks/request"
	stubs "github.com/pcarranza/meeseeks-box/testingstubs"
)

var shellCommand = shell.New(shell.CommandOpts{
	Cmd:  "echo",
	Help: "command that prints back the arguments passed",
})

func TestShellCommand(t *testing.T) {
	stubs.AssertEquals(t, "echo", shellCommand.Cmd())
	stubs.AssertEquals(t, []string{}, shellCommand.Args())
	stubs.AssertEquals(t, []string{}, shellCommand.AllowedGroups())
	stubs.AssertEquals(t, true, shellCommand.HasHandshake())
	stubs.AssertEquals(t, true, shellCommand.Record())
	stubs.AssertEquals(t, map[string]string{}, shellCommand.Templates())
	stubs.AssertEquals(t, command.DefaultCommandTimeout, shellCommand.Timeout())
	stubs.AssertEquals(t, "command that prints back the arguments passed", shellCommand.Help())
}

func TestExecuteShell(t *testing.T) {
	out, err := shellCommand.Execute(jobs.Job{
		Request: request.Request{Args: []string{"hello", "meeseeks\nsecond line"}},
	})
	stubs.Must(t, "failed to execute echo command", err)
	stubs.AssertEquals(t, "hello meeseeks\nsecond line\n", out)
}
