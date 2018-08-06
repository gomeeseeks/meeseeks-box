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
	mocks.Must(t, "could not add test command", commands.Add(commands.CommandRegistration{
		Name: "test",
		Cmd:  cmd,
	}))

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

func TestLoadingBuiltins(t *testing.T) {
	commands.Reset()

	_, ok := commands.Find(&meeseeks.Request{
		Command: "help",
	})
	mocks.AssertEquals(t, false, ok)

	commands.LoadBuiltins()

	defer commands.Reset()

	_, ok = commands.Find(&meeseeks.Request{
		Command: "help",
	})
	mocks.AssertEquals(t, true, ok)
}

func TestAddExistingFailsButReplaceWorks(t *testing.T) {
	commands.Reset()

	cmd := shell.New(meeseeks.CommandOpts{
		Cmd:  "echo",
		Help: meeseeks.NewHelp("echo"),
	})
	cmd2 := shell.New(meeseeks.CommandOpts{
		Cmd:  "echo2",
		Help: meeseeks.NewHelp("echo"),
	})

	mocks.Must(t, "could not add test command", commands.Add(commands.CommandRegistration{
		Name: "test",
		Cmd:  cmd,
	}))
	err := commands.Add(commands.CommandRegistration{
		Name: "test",
		Cmd:  cmd2,
	})
	mocks.AssertEquals(t, "command test is already registered", err.Error())

	commands.Replace(commands.CommandRegistration{
		Name: "test",
		Cmd:  cmd2,
	})
	c, ok := commands.Find(&meeseeks.Request{
		Command: "test",
	})

	mocks.AssertEquals(t, true, ok)
	mocks.AssertEquals(t, cmd2, c)

	commands.Reset()

	commands.Replace(commands.CommandRegistration{
		Name: "test",
		Cmd:  cmd2,
	})
}
