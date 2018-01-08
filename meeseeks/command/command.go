package command

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/command/parser"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/message"
)

// Command Errors
var (
	ErrCommandNotFound = fmt.Errorf("Can't find command")
	ErrNoCommandToRun  = errors.New("No command to run")
)

// Request is a structure that holds all the command execution request
type Request struct {
	Command  string
	Args     []string
	Username string
	Channel  string
	IsIM     bool
}

// RequestFromMessage gets a message and generates a valid request from it
func RequestFromMessage(msg message.Message) (Request, error) {
	args, err := parser.Parse(msg.GetText())
	if err != nil {
		return Request{}, err
	}

	if len(args) == 0 {
		return Request{}, ErrNoCommandToRun
	}

	return Request{
		Command:  args[0],
		Args:     args[1:],
		Username: msg.GetUsername(),
		Channel:  msg.GetChannel(),
		IsIM:     msg.IsIM(),
	}, nil
}

// Command is the base interface for any command
type Command interface {
	Execute(args ...string) (string, error)
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
	commands[config.BuiltinHelpCommand] = helpCommand{
		commands: &commands,
		Help:     "prints all the kwnown commands and its associated help",
	}
	commands[config.BuiltinVersionCommand] = versionCommand{
		Help: "prints the running meeseeks version",
	}
	commands[config.BuiltinGroupsCommand] = groupsCommand{
		Help: "prints the configured groups",
	}

	for name, configCommand := range cnf.Commands {
		command, err := buildCommand(configCommand)
		if err != nil {
			return Commands{}, err
		}
		commands[name] = command
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
func (c shellCommand) Execute(args ...string) (string, error) {
	cnfCommand := c.ConfiguredCommand()
	cmdArgs := append(cnfCommand.Args, args...)

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
