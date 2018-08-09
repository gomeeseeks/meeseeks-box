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
		commands.RegistrationArgs{
			Kind:   commands.KindLocalCommand,
			Action: commands.ActionRegister,
			Commands: []commands.CommandRegistration{
				commands.CommandRegistration{
					Name: "test",
					Cmd:  echoCmd,
				}}}))
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
		commands.RegistrationArgs{
			Action: commands.ActionRegister,
			Commands: []commands.CommandRegistration{
				commands.CommandRegistration{
					Name: "test",
					Cmd:  echoCmd,
				}}})), "Invalid registration, it has no kind")
	mocks.AssertEquals(t, fmt.Sprintf("%s", commands.Register(
		commands.RegistrationArgs{
			Action: commands.ActionRegister,
			Kind:   commands.KindLocalCommand,
			Commands: []commands.CommandRegistration{
				commands.CommandRegistration{
					Cmd: echoCmd,
				}}})), "Invalid command, it has no name")
	mocks.AssertEquals(t, fmt.Sprintf("%s", commands.Register(
		commands.RegistrationArgs{
			Action: commands.ActionRegister,
			Kind:   commands.KindLocalCommand,
			Commands: []commands.CommandRegistration{
				commands.CommandRegistration{
					Name: "test",
				}}})), "Invalid command test, it has no cmd")
	mocks.AssertEquals(t, fmt.Sprintf("%s", commands.Register(
		commands.RegistrationArgs{
			Action: "whatever",
			Kind:   commands.KindLocalCommand,
			Commands: []commands.CommandRegistration{
				commands.CommandRegistration{
					Name: "test",
					Cmd:  echoCmd,
				}}})), "Invalid action whatever")
	mocks.AssertEquals(t, fmt.Sprintf("%s", commands.Register(
		commands.RegistrationArgs{
			Action: commands.ActionRegister,
			Kind:   "whatever",
			Commands: []commands.CommandRegistration{
				commands.CommandRegistration{
					Name: "test",
					Cmd:  echoCmd,
				}}})), "Invalid kind of registration: whatever")
}

func TestReRegisteringChangingKindFails(t *testing.T) {
	mocks.Must(t, "could not register echo command", commands.Register(
		commands.RegistrationArgs{
			Kind:   commands.KindLocalCommand,
			Action: commands.ActionRegister,
			Commands: []commands.CommandRegistration{
				commands.CommandRegistration{
					Name: "echo",
					Cmd:  echoCmd,
				}}}))
	defer commands.Unregister("echo")

	mocks.AssertEquals(t, fmt.Sprintf("%s", commands.Register(
		commands.RegistrationArgs{
			Kind:   commands.KindRemoteCommand,
			Action: commands.ActionRegister,
			Commands: []commands.CommandRegistration{
				commands.CommandRegistration{
					Name: "echo",
					Cmd:  echoCmd,
				}}})), "incompatible command kind for an already known command")
}

func TestReRegisteringRemoteCommandsFails(t *testing.T) {
	mocks.Must(t, "could not register echo command", commands.Register(
		commands.RegistrationArgs{
			Kind:   commands.KindRemoteCommand,
			Action: commands.ActionRegister,
			Commands: []commands.CommandRegistration{
				commands.CommandRegistration{
					Name: "echo",
					Cmd:  echoCmd,
				}}}))
	defer commands.Unregister("echo")

	mocks.AssertEquals(t, fmt.Sprintf("%s", commands.Register(
		commands.RegistrationArgs{
			Kind:   commands.KindRemoteCommand,
			Action: commands.ActionRegister,
			Commands: []commands.CommandRegistration{
				commands.CommandRegistration{
					Name: "echo",
					Cmd:  echoCmd,
				}}})), "command echo is invalid, re-registering remote commands is not allowed yet")
}
