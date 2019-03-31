package shell_test

import (
	"context"
	"testing"
	"time"

	"gitlab.com/yakshaving.art/meeseeks-box/commands/shell"
	"gitlab.com/yakshaving.art/meeseeks-box/meeseeks"
	"gitlab.com/yakshaving.art/meeseeks-box/mocks"
)

var echoCommand = shell.New(meeseeks.CommandOpts{
	Cmd:  "echo",
	Help: meeseeks.NewHelp("command that prints back the arguments passed"),
})

var failCommand = shell.New(meeseeks.CommandOpts{
	Cmd:  "false",
	Help: meeseeks.NewHelp("command that fails"),
})

var sleepCommand = shell.New(meeseeks.CommandOpts{
	Cmd:  "sleep",
	Args: []string{"10"},
	Help: meeseeks.NewHelp("command that sleeps"),
})

func TestShellCommand(t *testing.T) {
	mocks.AssertEquals(t, "echo", echoCommand.GetCmd())
	mocks.AssertEquals(t, []string{}, echoCommand.GetArgs())
	mocks.AssertEquals(t, []string{}, echoCommand.GetAllowedGroups())
	mocks.AssertEquals(t, []string{}, echoCommand.GetAllowedChannels())
	mocks.AssertEquals(t, false, echoCommand.HasHandshake())
	mocks.AssertEquals(t, true, echoCommand.MustRecord())
	mocks.AssertEquals(t, meeseeks.DefaultCommandTimeout, echoCommand.GetTimeout())
	mocks.AssertEquals(t, "command that prints back the arguments passed", echoCommand.GetHelp().GetSummary())
	mocks.AssertEquals(t, []string{}, echoCommand.GetHelp().GetArgs())
}

func TestExecuteEcho(t *testing.T) {
	mocks.WithTmpDB(func(_ string) {
		out, err := echoCommand.Execute(context.Background(), meeseeks.Job{
			ID:      1,
			Request: meeseeks.Request{Args: []string{"hello", "meeseeks\nsecond line"}},
		})
		mocks.Must(t, "failed to execute echo command", err)
		mocks.AssertEquals(t, "hello meeseeks\nsecond line\n", out)
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
		mocks.AssertEquals(t, "context canceled", err.Error())
	})
}
