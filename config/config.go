package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"time"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// Authorization Strategies determine who has access to what
const (
	AuthStrategyAny      = "any"
	AuthStrategyUserList = "userlist"
)

// Default messages to use
var (
	DefaultHandshake      = []string{"I'm Mr Meeseeks! look at me!", "Mr Meeseeks!", "Uuuuh, yeah! can do!", "Can doo...", "Uuuuh, ok!"}
	DefaultSuccess        = []string{"All done!", "Mr Meeseeks", "Uuuuh, nice!"}
	DefaultFailed         = []string{"Uuuh!, no, it failed"}
	DefaultUnauthorized   = []string{"Uuuuh, yeah! you are not allowed to do"}
	DefaultUnknownCommand = []string{"Uuuh! no, I don't know how to do"}
)

// Defaults for commands
var (
	DefaultCommandTimeout = 60 * time.Second
)

// Builtin Commands
var builtinCommands = map[string]Command{
	"echo": Command{
		Cmd:          "echo",
		Timeout:      5 * time.Second,
		AuthStrategy: AuthStrategyAny,
	},
}

// New parses the configuration from a reader into an object and returns it
func New(r io.Reader) (Config, error) {
	c := Config{
		Messages: map[string][]string{
			"handshake":      DefaultHandshake,
			"success":        DefaultSuccess,
			"failed":         DefaultFailed,
			"unauthorized":   DefaultUnauthorized,
			"unknowncommand": DefaultUnknownCommand,
		},
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return c, fmt.Errorf("could not read configuration: %s", err)
	}

	err = yaml.Unmarshal(b, &c)
	if err != nil {
		return c, fmt.Errorf("could not parse configuration: %s", err)
	}

	for name, command := range c.Commands {
		if command.AuthStrategy == "" {
			log.Debugf("Applying default AuthStrategy %s to command %s", AuthStrategyAny, name)
			command.AuthStrategy = AuthStrategyAny
		}
		if command.Timeout == 0 {
			log.Debugf("Applying default Timeout %d to command %s", DefaultCommandTimeout, name)
			command.Timeout = DefaultCommandTimeout
		}
	}

	return c, nil
}

// Config is the struct used to load MrMeeseeks configuration yaml
type Config struct {
	Messages map[string][]string `yaml:"messages"`
	Commands map[string]Command  `yaml:"commands"`
}

// Command is the struct that handles a command configuration
type Command struct {
	Cmd          string          `yaml:"command"`
	Args         []string        `yaml:"arguments"`
	Authorized   []string        `yaml:"authorized"`
	AuthStrategy string          `yaml:"auth_strategy"`
	Timeout      time.Duration   `yaml:"timeout"`
	Templates    CommandTemplate `yaml:"templates"`
}

// CommandTemplate is the struct in which the templates used to render a command are kept
type CommandTemplate struct {
	Handshake string `yaml:"on_handshake"`
	Success   string `yaml:"on_success"`
	Failure   string `yaml:"on_failure"`
}

// GetCommands builds the definitive command list
func (c Config) GetCommands() map[string]Command {
	commands := make(map[string]Command)
	for name, command := range builtinCommands {
		commands[name] = command
	}
	for name, command := range c.Commands {
		if _, ok := commands[name]; ok {
			log.Infof("Shadowing builtin command %s on configuration", name)
		}
		commands[name] = command
	}
	return commands
}
