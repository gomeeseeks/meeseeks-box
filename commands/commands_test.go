package commands_test

import (
	"testing"

	"github.com/gomeeseeks/meeseeks-box/commands"
	"github.com/gomeeseeks/meeseeks-box/commands/shell"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/mocks"
)

func TestAdAndFindCommands(t *testing.T) {
	cmd := shell.New(meeseeks.CommandOpts{
		Cmd:  "echo",
		Help: meeseeks.NewHelp("echo"),
	})
	mocks.Must(t, "could not add test command", commands.Add(commands.NewLocalCommand(
		"test",
		cmd,
	)))

	c, ok := commands.Find(&meeseeks.Request{
		Command: "test",
	})

	mocks.AssertEquals(t, true, ok)
	mocks.AssertEquals(t, cmd, c)

	commands.Remove("test")

	_, ok = commands.Find(&meeseeks.Request{
		Command: "test",
	})
	mocks.AssertEquals(t, false, ok)

	commands.Remove("test")
}
