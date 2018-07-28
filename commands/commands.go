package commands

import (
	"fmt"
	"sync"

	"github.com/gomeeseeks/meeseeks-box/commands/builtins"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/metrics"
	"github.com/gomeeseeks/meeseeks-box/persistence"
	"github.com/sirupsen/logrus"
)

var commands map[string]meeseeks.Command
var mutex sync.Mutex

func init() {
	Reset()
}

// Reset flushes all the commands and loads only the builtins
func Reset() {
	mutex.Lock()
	defer mutex.Unlock()

	commands = make(map[string]meeseeks.Command)
}

// LoadBuiltins loads the builtin commands
func LoadBuiltins() {
	mutex.Lock()
	defer mutex.Unlock()

	for name, cmd := range builtins.Commands {
		commands[name] = cmd
	}
	builtins.AddHelpCommand(commands)
}

// CommandRegistration is used to register a new command in the commands map
type CommandRegistration struct {
	Name string
	Cmd  meeseeks.Command
}

// Add adds a new command to the map
func Add(cmds ...CommandRegistration) error {
	mutex.Lock()
	defer mutex.Unlock()

	for _, cmd := range cmds {
		if _, ok := commands[cmd.Name]; ok {
			return fmt.Errorf("command %s is already registered", cmd.Name)
		}
	}

	for _, cmd := range cmds {
		commands[cmd.Name] = cmd.Cmd
	}
	return nil
}

// Replace replaces an already registered command
func Replace(cmd CommandRegistration) {
	if _, ok := commands[cmd.Name]; !ok {
		logrus.Infof("command %s not found for replacing", cmd.Name)
	}
	commands[cmd.Name] = cmd.Cmd
}

// Find looks up the given command by name and returns.
//
// This method implements the map interface as in returning true of false in the
// case the command exists in the map
func Find(req *meeseeks.Request) (meeseeks.Command, bool) {
	aliasedCommand, args, err := persistence.Aliases().Get(req.UserID, req.Command)
	if err != nil {
		logrus.Debugf("Failed to get alias %s", req.Command)
	}

	if cmd, ok := commands[aliasedCommand]; ok {
		logrus.Infof("Command %s is an alias for %s", req.Command, aliasedCommand)
		metrics.AliasedCommandsCount.Inc()
		req.Command = aliasedCommand
		req.Args = append(args, req.Args...)

		return cmd, ok
	}

	cmd, ok := commands[req.Command]
	return cmd, ok
}
