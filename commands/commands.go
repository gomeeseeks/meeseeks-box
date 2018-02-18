package commands

import (
	"sync"

	"github.com/gomeeseeks/meeseeks-box/aliases"
	"github.com/gomeeseeks/meeseeks-box/command"
	"github.com/gomeeseeks/meeseeks-box/commands/builtins"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/sirupsen/logrus"
)

var commands map[string]command.Command
var mutex sync.Mutex

func init() {
	Reset()
}

// Reset flushes all the commands and loads only the builtins
func Reset() {
	mutex.Lock()
	defer mutex.Unlock()

	commands = make(map[string]command.Command)
	for name, cmd := range builtins.Commands {
		commands[name] = cmd
	}

	builtins.AddHelpCommand(commands)
}

// Add adds a new command to the map
func Add(name string, cmd command.Command) {
	mutex.Lock()
	defer mutex.Unlock()

	commands[name] = cmd
}

// Find looks up the given command by name and returns.
//
// This method implements the map interface as in returning true of false in the
// case the command exists in the map
func Find(req *meeseeks.Request) (command.Command, bool) {
	aliasedCommand, args, err := aliases.Get(req.UserID, req.Command)
	if err != nil {
		logrus.Debugf("Failed to get alias %s", req.Command)
	}

	if cmd, ok := commands[aliasedCommand]; ok {
		logrus.Infof("Command %s is an alias", req.Command)
		req.Args = append(args, req.Args...)

		return cmd, ok
	}

	cmd, ok := commands[req.Command]
	return cmd, ok
}
