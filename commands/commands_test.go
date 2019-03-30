package commands_test

import (
	"fmt"
	"testing"

	"gitlab.com/yakshaving.art/meeseeks-box/commands"
	"gitlab.com/yakshaving.art/meeseeks-box/commands/shell"
	"gitlab.com/yakshaving.art/meeseeks-box/meeseeks"
	"gitlab.com/yakshaving.art/meeseeks-box/mocks"
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

	mocks.Must(t, "could not unregister test command", commands.Register(
		commands.RegistrationArgs{
			Kind:   commands.KindLocalCommand,
			Action: commands.ActionUnregister,
			Commands: []commands.CommandRegistration{
				commands.CommandRegistration{
					Name: "test",
					Cmd:  echoCmd,
				}}}))

	_, ok = commands.Find(&meeseeks.Request{
		Command: "test",
	})
	mocks.AssertEquals(t, false, ok)

	mocks.AssertEquals(t, "can't unregister an unknown command: test", fmt.Sprintf("%s", commands.Register(
		commands.RegistrationArgs{
			Kind:   commands.KindLocalCommand,
			Action: commands.ActionUnregister,
			Commands: []commands.CommandRegistration{
				commands.CommandRegistration{
					Name: "test",
					Cmd:  echoCmd,
				}}})))
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
	defer func() {
		mocks.Must(t, "could not unregister test command", commands.Register(
			commands.RegistrationArgs{
				Kind:   commands.KindLocalCommand,
				Action: commands.ActionUnregister,
				Commands: []commands.CommandRegistration{
					commands.CommandRegistration{
						Name: "echo",
						Cmd:  echoCmd,
					}}}))
	}()

	mocks.AssertEquals(t, fmt.Sprintf("%s", commands.Register(
		commands.RegistrationArgs{
			Kind:   commands.KindRemoteCommand,
			Action: commands.ActionRegister,
			Commands: []commands.CommandRegistration{
				commands.CommandRegistration{
					Name: "echo",
					Cmd:  echoCmd,
				}}})), "incompatible command kind for an already known command: echo")
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

	mocks.AssertEquals(t, fmt.Sprintf("%s", commands.Register(
		commands.RegistrationArgs{
			Kind:   commands.KindRemoteCommand,
			Action: commands.ActionRegister,
			Commands: []commands.CommandRegistration{
				commands.CommandRegistration{
					Name: "echo",
					Cmd:  echoCmd,
				}}})), "command echo is invalid, re-registering remote commands is not allowed yet")

	mocks.Must(t, "could not unregister echo command", commands.Register(
		commands.RegistrationArgs{
			Kind:   commands.KindRemoteCommand,
			Action: commands.ActionUnregister,
			Commands: []commands.CommandRegistration{
				commands.CommandRegistration{
					Name: "echo",
					Cmd:  echoCmd,
				}}}))
}
