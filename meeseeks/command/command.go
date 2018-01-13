package command

import (
	"context"
	"fmt"
	"os/exec"

	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/request"

	"gitlab.com/mr-meeseeks/meeseeks-box/config"
)

// Command Errors
var (
	ErrCommandNotFound = fmt.Errorf("Can't find command")
)

// Command is the base interface for any command
type Command interface {
	Execute(req request.Request) (string, error)
	HasHandshake() bool
	ConfiguredCommand() config.Command
}

// Commands holds the final set of configured commands
type Commands struct {
	commands map[string]Command
}

// New builds a new commands based on a configuration
func New(cnf config.Config) (Commands, error) {
	// Add builtin commands
	commands := make(map[string]Command)
	for name, command := range builtInCommands {
		commands[name] = command
	}

	for name, configCommand := range cnf.Commands {
		command, err := buildCommand(configCommand)
		if err != nil {
			return Commands{}, err
		}
		commands[name] = command
	}

	commands[BuiltinHelpCommand] = helpCommand{
		commands: &commands,
		Help:     "prints all the kwnown commands and its associated help",
	}

	return Commands{
		commands: commands,
	}, nil
}

// Find looks up a command by name and returns it or an error
func (c Commands) Find(name string) (Command, error) {
	cmd, ok := c.commands[name]
	if !ok {
		return nil, ErrCommandNotFound
	}
	return cmd, nil
}

// buildCommand creates a command instance based on the configuration
func buildCommand(cmd config.Command) (Command, error) {
	switch cmd.Type {
	case config.ShellCommandType:
		return newShellCommand(cmd)
	}
	return nil, ErrCommandNotFound
}

// ShellCommand is a command that will be executed locally through an exec.CommandContext
type shellCommand struct {
	config.Command
}

// NewShellCommand returns a new command that is executed inside a shell
func newShellCommand(cmd config.Command) (Command, error) {
	return shellCommand{
		Command: cmd,
	}, nil
}

// Execute implements Command.Execute for the ShellCommand
func (c shellCommand) Execute(req request.Request) (string, error) {
	cnfCommand := c.ConfiguredCommand()
	cmdArgs := append(cnfCommand.Args, req.Args...)

	ctx, cancelFunc := context.WithTimeout(context.Background(), cnfCommand.Timeout)
	defer cancelFunc()

	shellCommand := exec.CommandContext(ctx, cnfCommand.Cmd, cmdArgs...)
	out, err := shellCommand.CombinedOutput()

	return string(out), err
}

func (c shellCommand) HasHandshake() bool {
	return true
}

func (c shellCommand) ConfiguredCommand() config.Command {
	return c.Command
}

// Help returns the help from the configured command
func (c shellCommand) Help() string {
	return c.Command.Help
}
