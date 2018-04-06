package shell_test

import (
	"context"
	"testing"
	"time"

	"github.com/gomeeseeks/meeseeks-box/commands/shell"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/mocks"
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
	mocks.AssertEquals(t, "echo", echoCommand.Cmd())
	mocks.AssertEquals(t, []string{}, echoCommand.Args())
	mocks.AssertEquals(t, []string{}, echoCommand.AllowedGroups())
	mocks.AssertEquals(t, []string{}, echoCommand.AllowedChannels())
	mocks.AssertEquals(t, true, echoCommand.HasHandshake())
	mocks.AssertEquals(t, true, echoCommand.Record())
	mocks.AssertEquals(t, map[string]string{}, echoCommand.Templates())
	mocks.AssertEquals(t, meeseeks.DefaultCommandTimeout, echoCommand.Timeout())
	mocks.AssertEquals(t, "command that prints back the arguments passed", echoCommand.Help().GetSummary())
	mocks.AssertEquals(t, []string{}, echoCommand.Help().GetArgs())
}

func TestExecuteEcho(t *testing.T) {
	mocks.WithTmpDB(func(_ string) {
		out, err := echoCommand.Execute(context.Background(), meeseeks.Job{
			ID:      1,
			Request: meeseeks.Request{Args: []string{"hello", "meeseeks\nsecond line"}},
		})
		mocks.Must(t, "failed to execute echo command", err)
		mocks.AssertEquals(t, "hello meeseeks\nsecond line", out)
	})
}

func TestExecuteFail(t *testing.T) {
	mocks.WithTmpDB(func(_ string) {
		_, err := failCommand.Execute(context.Background(), meeseeks.Job{
			ID:      2,
			Request: meeseeks.Request{},
		})
		mocks.AssertEquals(t, "exit status 1", err.Error())
	})
}

func TestSleepingCanBeWokenUp(t *testing.T) {
	mocks.WithTmpDB(func(_ string) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-time.After(10 * time.Millisecond)
			cancel()
		}()
		_, err := sleepCommand.Execute(ctx, meeseeks.Job{
			ID:      3,
			Request: meeseeks.Request{},
		})
		mocks.AssertEquals(t, "signal: killed", err.Error())
	})
}
