package commands

import (
	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	"gitlab.com/mr-meeseeks/meeseeks-box/version"
)

type builtinCommand struct {
}

var allowAllConfiguredCommand = config.Command{
	AuthStrategy: config.AuthStrategyAny,
}

func (b builtinCommand) HasHandshake() bool {
	return false
}

func (b builtinCommand) ConfiguredCommand() config.Command {
	return allowAllConfiguredCommand
}

type versionCommand struct {
	builtinCommand
}

func (v versionCommand) Execute(args ...string) (string, error) {
	return version.AppVersion, nil
}

type helpCommand struct {
	builtinCommand
	commands *map[string]Command
}

func (h helpCommand) Execute(args ...string) (string, error) {
	return "invalid", nil
}
