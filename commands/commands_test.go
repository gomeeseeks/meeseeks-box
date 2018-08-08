package commands_test

import (
	"fmt"
	"testing"

	"github.com/gomeeseeks/meeseeks-box/commands"
	"github.com/gomeeseeks/meeseeks-box/commands/shell"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/mocks"
)

var echoCmd = shell.New(meeseeks.CommandOpts{
	Cmd:  "echo",
	Help: meeseeks.NewHelp("echo"),
})

func TestAddAndFindCommands(t *testing.T) {
	mocks.Must(t, "could not add test command", commands.Register(
		commands.CommandRegistration{
			Name: "test",
			Cmd:  echoCmd,
			Kind: commands.KindLocalCommand,
		}))
	defer commands.Unregister("test")

	c, ok := commands.Find(&meeseeks.Request{
		Command: "test",
	})

	mocks.AssertEquals(t, true, ok)
	mocks.AssertEquals(t, echoCmd, c)

	cmds := commands.All()
	c, ok = cmds["test"]
	mocks.AssertEquals(t, true, ok)
	mocks.AssertEquals(t, echoCmd, c)
	mocks.AssertEquals(t, 1, len(cmds))

	commands.Unregister("test")

	_, ok = commands.Find(&meeseeks.Request{
		Command: "test",
	})
	mocks.AssertEquals(t, false, ok)

	commands.Unregister("test")
}

func TestAddingAnInvalidCommandFails(t *testing.T) {
	mocks.AssertEquals(t, fmt.Sprintf("%s", commands.Register(
		commands.CommandRegistration{
			Name: "test",
			Cmd:  echoCmd,
		})), "Invalid command test, it has no kind")
	mocks.AssertEquals(t, fmt.Sprintf("%s", commands.Register(
		commands.CommandRegistration{
			Cmd:  echoCmd,
			Kind: commands.KindLocalCommand,
		})), "Invalid command, it has no name")
	mocks.AssertEquals(t, fmt.Sprintf("%s", commands.Register(
		commands.CommandRegistration{
			Name: "test",
			Kind: commands.KindLocalCommand,
		})), "Invalid command test, it has no cmd")
}

func TestReRegisteringChangingKindFails(t *testing.T) {
	mocks.Must(t, "could not register echo command", commands.Register(
		commands.CommandRegistration{
			Name: "echo",
			Cmd:  echoCmd,
			Kind: commands.KindLocalCommand,
		}))
	defer commands.Unregister("echo")

	mocks.AssertEquals(t, fmt.Sprintf("%s", commands.Register(
		commands.CommandRegistration{
			Name: "echo",
			Cmd:  echoCmd,
			Kind: commands.KindRemoteCommand,
		})), "Command echo would change the kind from local to remote, this is not allowed")
}
