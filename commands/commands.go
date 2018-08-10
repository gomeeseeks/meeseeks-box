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

// RegistrationArgs allows to register new commands
type RegistrationArgs struct {
	Kind     string
	Action   string
	Commands []CommandRegistration
}

func (r RegistrationArgs) validate() error {
	if strings.TrimSpace(r.Kind) == "" {
		return fmt.Errorf("Invalid registration, it has no kind")
	}
	switch r.Kind {
	case KindBuiltinCommand, KindLocalCommand, KindRemoteCommand:
		break
	default:
		return fmt.Errorf("Invalid kind of registration: %s", r.Kind)
	}

	switch r.Action {
	case ActionRegister, ActionUnregister:
		break
	default:
		return fmt.Errorf("Invalid action %s", r.Action)
	}

	for _, cmd := range r.Commands {
		if err := cmd.validate(); err != nil {
			return err
		}

		if knownCommand, ok := commands[cmd.Name]; ok {
			if knownCommand.kind != r.Kind {
				return fmt.Errorf("incompatible command kind for an already known command")
			}
			if knownCommand.kind == KindRemoteCommand {
				return fmt.Errorf("command %s is invalid, re-registering remote commands is not allowed yet",
					cmd.Name)
			}
		} else {
			if r.Action == ActionUnregister {
				return fmt.Errorf("can't unregister a non registered command")
			}
		}
	}

	return nil
}

func (r RegistrationArgs) process() {
	switch r.Action {
	case ActionRegister:
		r.registerCommands()
	default:
		r.unregisterCommands()
	}
}

func (r RegistrationArgs) unregisterCommands() {
	for _, cmd := range r.Commands {
		delete(commands, cmd.Name)
	}
}

func (r RegistrationArgs) registerCommands() {
	for _, cmd := range r.Commands {
		commands[cmd.Name] = commandHub{
			cmd:  cmd.Cmd,
			kind: r.Kind,
		}
	}
}

// CommandRegistration is used to register a new command in the commands map
type CommandRegistration struct {
	Name string
	Cmd  meeseeks.Command
}

func (c CommandRegistration) validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("Invalid command, it has no name")
	}
	if c.Cmd == nil {
		return fmt.Errorf("Invalid command %s, it has no cmd", c.Name)
	}

	return nil
}

// Register registers all the commands passed if they are valid
func Register(args RegistrationArgs) error {
	mutex.Lock()
	defer mutex.Unlock()

	if err := args.validate(); err != nil {
		return err
	}

	args.process()

	return nil
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
