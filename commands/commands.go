package commands

import (
	"fmt"
	"strings"
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

// Kind of commands we can register
const (
	KindLocalCommand   = "local"
	KindRemoteCommand  = "remote"
	KindBuiltinCommand = "builtin"
)

// Action to perform when dealing with commands
const (
	ActionRegister   = "register"
	ActionUnregister = "unregister"
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
	Name   string
	Cmd    meeseeks.Command
	Kind   string
	Action string
}

func (c CommandRegistration) validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("Invalid command, it has no name")
	}
	if c.Cmd == nil {
		return fmt.Errorf("Invalid command %s, it has no cmd", c.Name)
	}
	if strings.TrimSpace(c.Kind) == "" {
		return fmt.Errorf("Invalid command %s, it has no kind", c.Name)
	}

	switch c.Kind {
	case KindBuiltinCommand, KindLocalCommand, KindRemoteCommand:
		break
	default:
		return fmt.Errorf("Invalid kind of command: %s", c.Kind)
	}

	switch c.Action {
	case ActionRegister, ActionUnregister:
		break
	default:
		return fmt.Errorf("Invalid action %s", c.Action)
	}

	return nil
}

// Register registers all the commands passed if they are valid
func Register(cmds ...CommandRegistration) error {
	mutex.Lock()
	defer mutex.Unlock()

	for _, cmd := range cmds {
		if err := cmd.validate(); err != nil {
			return err
		}

		if knownCommand, ok := commands[cmd.Name]; ok {
			if knownCommand.kind != cmd.Kind {
				return fmt.Errorf("command %s would change the kind from %s to %s, this is not allowed",
					cmd.Name, knownCommand.kind, cmd.Kind)
			}
			if cmd.Kind == KindRemoteCommand {
				return fmt.Errorf("command %s is invalid, replacing remote commands is not allowed yet",
					cmd.Name)
			}
		}
	}

	logrus.Debugf("appending commands %#v", cmds)
	for _, cmd := range cmds {
		commands[cmd.Name] = commandHub{
			cmd:  cmd.Cmd,
			kind: cmd.Kind,
		}
	}
	return nil
}

// Unregister unregisters commands from the registration list
func Unregister(cmds ...string) {
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
