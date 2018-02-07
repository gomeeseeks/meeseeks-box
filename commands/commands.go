package commands

import (
	"sync"

	"github.com/gomeeseeks/meeseeks-box/command"
	"github.com/gomeeseeks/meeseeks-box/commands/builtins"
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

// Find looks up the given command by name and returns.
//
// This method implements the map interface as in returning true of false in the
// case the command exists in the map
func Find(name string) (command.Command, bool) {
	cmd, ok := commands[name]
	return cmd, ok
}

// Add adds a new command to the map
func Add(name string, cmd command.Command) {
	mutex.Lock()
	defer mutex.Unlock()

	commands[name] = cmd
}
