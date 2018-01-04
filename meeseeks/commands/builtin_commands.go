package commands

import (
	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	"gitlab.com/mr-meeseeks/meeseeks-box/version"
)

var builtinCommands = map[string]Command{
	config.BuiltinCommandVersion: versionCommand{},
}

func newBuiltinCommand(cmd config.Command) (Command, error) {
	if command, ok := builtinCommands[cmd.Cmd]; ok {
		return command, nil
	}
	return nil, ErrCommandNotFound
}

type builtinCommand struct {
}

func (b builtinCommand) HasHandshake() bool {
	return false
}

type versionCommand struct {
	builtinCommand
}

func (v versionCommand) Execute(args ...string) (string, error) {
	return version.AppVersion, nil
}
