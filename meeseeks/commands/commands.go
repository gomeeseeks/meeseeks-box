package commands

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"gitlab.com/mr-meeseeks/meeseeks-box/config"
)

// Command Errors
var (
	ErrCommandNotFound = fmt.Errorf("Can't find command")
)

// Command is the base interface for any command
type Command interface {
	Execute(args ...string) (string, error)
	HasHandshake() bool
}

// New builds a new command from a configured one
func New(cmd config.Command) (Command, error) {
	switch cmd.Type {
	case config.ShellCommandType:
		return newShellCommand(cmd)

	case config.BuiltinCommandType:
		return newBuiltinCommand(cmd)

	default:
		return nil, ErrCommandNotFound

	}
}

// ShellCommand is a command that will be executed locally through an exec.CommandContext
type shellCommand struct {
	timeout time.Duration
	cmd     string
	args    []string
}

// NewShellCommand returns a new command that is executed inside a shell
func newShellCommand(cmd config.Command) (Command, error) {
	timeout := time.Duration(cmd.Timeout) * time.Second
	if cmd.Timeout == 0 {
		timeout = config.DefaultCommandTimeout
	}

	return shellCommand{
		timeout: timeout,
		cmd:     cmd.Cmd,
		args:    cmd.Args,
	}, nil
}

// Execute implements Command.Execute for the ShellCommand
func (c shellCommand) Execute(args ...string) (string, error) {
	cmdArgs := append(c.args, args...)

	ctx, cancelFunc := context.WithTimeout(context.Background(), c.timeout)
	defer cancelFunc()

	shellCommand := exec.CommandContext(ctx, c.cmd, cmdArgs...)
	out, err := shellCommand.CombinedOutput()

	return string(out), err
}

func (c shellCommand) HasHandshake() bool {
	return true
}
