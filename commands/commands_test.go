package commands_test

import (
	"fmt"
	"testing"

	"github.com/gomeeseeks/meeseeks-box/commands"
	"github.com/gomeeseeks/meeseeks-box/commands/shell"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/mocks"
)

func TestAddAndFindCommands(t *testing.T) {
	cmd := shell.New(meeseeks.CommandOpts{
		Cmd:  "echo",
		Help: meeseeks.NewHelp("echo"),
	})
	mocks.Must(t, "could not add test command", commands.Register(
		commands.CommandRegistration{
			Name: "test",
			Cmd:  cmd,
			Kind: commands.KindLocalCommand,
		}))

	c, ok := commands.Find(&meeseeks.Request{
		Command: "test",
	})

	mocks.AssertEquals(t, true, ok)
	mocks.AssertEquals(t, cmd, c)

	// Try through using All, we should get a map with only this cmd
	cmds := commands.All()
	c, ok = cmds["test"]
	mocks.AssertEquals(t, true, ok)
	mocks.AssertEquals(t, cmd, c)
	mocks.AssertEquals(t, 1, len(cmds))

	mocks.AssertEquals(t, fmt.Sprintf("%s", commands.Register(
		commands.CommandRegistration{
			Name: "test",
			Cmd:  cmd,
			Kind: commands.KindLocalCommand,
		})), "command test is already registered")
	commands.Unregister("test")

	_, ok = commands.Find(&meeseeks.Request{
		Command: "test",
	})
	mocks.AssertEquals(t, false, ok)

	commands.Unregister("test")
}

func TestAddingAnInvalidCommandFails(t *testing.T) {
	cmd := shell.New(meeseeks.CommandOpts{
		Cmd:  "echo",
		Help: meeseeks.NewHelp("echo"),
	})
	mocks.AssertEquals(t, fmt.Sprintf("%s", commands.Register(
		commands.CommandRegistration{
			Name: "test",
			Cmd:  cmd,
		})), "Invalid command test, it has no kind")
	mocks.AssertEquals(t, fmt.Sprintf("%s", commands.Register(
		commands.CommandRegistration{
			Cmd:  cmd,
			Kind: commands.KindLocalCommand,
		})), "Invalid command, it has no name")
	mocks.AssertEquals(t, fmt.Sprintf("%s", commands.Register(
		commands.CommandRegistration{
			Name: "test",
			Kind: commands.KindLocalCommand,
		})), "Invalid command test, it has no cmd")

}
