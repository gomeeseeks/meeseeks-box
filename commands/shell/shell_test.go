package shell_test

import (
	"context"
	"testing"
	"time"

	"github.com/gomeeseeks/meeseeks-box/command"
	"github.com/gomeeseeks/meeseeks-box/commands/shell"
	"github.com/gomeeseeks/meeseeks-box/jobs"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/request"
	stubs "github.com/gomeeseeks/meeseeks-box/testingstubs"
)

var echoCommand = shell.New(shell.CommandOpts{
	Cmd:  "echo",
	Help: shell.NewHelp("command that prints back the arguments passed"),
})

var failCommand = shell.New(shell.CommandOpts{
	Cmd:  "false",
	Help: shell.NewHelp("command that fails"),
})

var sleepCommand = shell.New(shell.CommandOpts{
	Cmd:  "sleep",
	Args: []string{"10"},
	Help: shell.NewHelp("command that sleeps"),
})

func TestShellCommand(t *testing.T) {
	stubs.AssertEquals(t, "echo", echoCommand.Cmd())
	stubs.AssertEquals(t, []string{}, echoCommand.Args())
	stubs.AssertEquals(t, []string{}, echoCommand.AllowedGroups())
	stubs.AssertEquals(t, true, echoCommand.HasHandshake())
	stubs.AssertEquals(t, true, echoCommand.Record())
	stubs.AssertEquals(t, map[string]string{}, echoCommand.Templates())
	stubs.AssertEquals(t, command.DefaultCommandTimeout, echoCommand.Timeout())
	stubs.AssertEquals(t, "command that prints back the arguments passed", echoCommand.Help().GetSummary())
	stubs.AssertEquals(t, []string{}, echoCommand.Help().GetArgs())
}

func TestExecuteEcho(t *testing.T) {
	stubs.WithTmpDB(func(_ string) {
		out, err := echoCommand.Execute(context.Background(), jobs.Job{
			ID:      1,
			Request: request.Request{Args: []string{"hello", "meeseeks\nsecond line"}},
		})
		stubs.Must(t, "failed to execute echo command", err)
		stubs.AssertEquals(t, "hello meeseeks\nsecond line\n", out)
	})
}

func TestExecuteFail(t *testing.T) {
	stubs.WithTmpDB(func(_ string) {
		_, err := failCommand.Execute(context.Background(), jobs.Job{
			ID:      2,
			Request: request.Request{},
		})
		stubs.AssertEquals(t, "exit status 1", err.Error())
	})
}

func TestSleepingCanBeWokenUp(t *testing.T) {
	stubs.WithTmpDB(func(_ string) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-time.After(10 * time.Millisecond)
			cancel()
		}()
		_, err := sleepCommand.Execute(ctx, jobs.Job{
			ID:      3,
			Request: request.Request{},
		})
		stubs.AssertEquals(t, "signal: killed", err.Error())
	})
}
