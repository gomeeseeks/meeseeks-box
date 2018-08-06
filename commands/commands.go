package commands

import (
	"fmt"
	"sync"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/metrics"
	"github.com/gomeeseeks/meeseeks-box/persistence"
	"github.com/sirupsen/logrus"
)

func init() {
	Reset()
}

var commands map[string]commandHub
var mutex sync.Mutex

// Reset flushes all the commands
func Reset() {
	mutex.Lock()
	defer mutex.Unlock()

	commands = make(map[string]commandHub)
}

const (
	kindLocalCommand   = "local"
	kindRemoteCommand  = "remote"
	kindBuiltinCommand = "builtin"
)

type commandHub struct {
	kind string
	cmd  meeseeks.Command
}

// All returns all the currently registered commands
func All() map[string]meeseeks.Command {
	c := make(map[string]meeseeks.Command)
	for name, hub := range commands {
		c[name] = hub.cmd
	}
	return c
}

// CommandRegistration is used to register a new command in the commands map
type CommandRegistration struct {
	name string
	cmd  meeseeks.Command
	kind string
}

// NewBuiltinCommand creates a new local command
func NewBuiltinCommand(name string, cmd meeseeks.Command) CommandRegistration {
	return CommandRegistration{
		name: name,
		cmd:  cmd,
		kind: kindLocalCommand,
	}
}

// NewLocalCommand creates a new local command
func NewLocalCommand(name string, cmd meeseeks.Command) CommandRegistration {
	return CommandRegistration{
		name: name,
		cmd:  cmd,
		kind: kindLocalCommand,
	}
}

// NewRemoteCommand creates a new local command
func NewRemoteCommand(name string, cmd meeseeks.Command) CommandRegistration {
	return CommandRegistration{
		name: name,
		cmd:  cmd,
		kind: kindRemoteCommand,
	}
}

// Add adds a new command to the map
func Add(cmds ...CommandRegistration) error {
	mutex.Lock()
	defer mutex.Unlock()

	for _, cmd := range cmds {
		if _, ok := commands[cmd.name]; ok {
			return fmt.Errorf("command %s is already registered", cmd.name)
		}
	}

	logrus.Debugf("appending commands %#v", cmds)

	for _, cmd := range cmds {
		commands[cmd.name] = commandHub{
			cmd:  cmd.cmd,
			kind: cmd.kind,
		}
	}
	return nil
}

// Remove unregisters commands from the registration list
func Remove(cmds ...string) {
	mutex.Lock()
	defer mutex.Unlock()

	for _, name := range cmds {
		if _, ok := commands[name]; ok {
			delete(commands, name)
		} else {
			logrus.Warnf("could not delete command %s because it's not to be found", name)
		}
	}
}

// Find looks up the given command by name and returns.
//
// This method implements the map interface as in returning true of false in the
// case the command exists in the map
func Find(req *meeseeks.Request) (meeseeks.Command, bool) {
	mutex.Lock()
	defer mutex.Unlock()

	aliasedCommand, args, err := persistence.Aliases().Get(req.UserID, req.Command)
	if err != nil {
		logrus.Debugf("Failed to get alias %s", req.Command)
	}

	if cmd, ok := commands[aliasedCommand]; ok {
		logrus.Infof("Command %s is an alias for %s", req.Command, aliasedCommand)
		metrics.AliasedCommandsCount.Inc()
		req.Command = aliasedCommand
		req.Args = append(args, req.Args...)

		return cmd.cmd, ok
	}

	cmd, ok := commands[req.Command]
	return cmd.cmd, ok
}
