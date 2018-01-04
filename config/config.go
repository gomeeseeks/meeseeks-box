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
	AuthStrategyNone     = "none"
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
const (
	DefaultCommandTimeout = 60 * time.Second
)

// Default colors
const (
	DefaultInfoColorMessage    = ""
	DefaultErrColorMessage     = "#cc3300"
	DefaultSuccessColorMessage = "#009900"
)

// Command types
const (
	BuiltinCommandType = iota
	ShellCommandType
	RemoteCommandType
)

// Builtin Commands Names
const (
	BuiltinVersionCommand = "version"
	BuiltinHelpCommand    = "help"
)

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
		Colors: MessageColors{
			Info:    DefaultInfoColorMessage,
			Success: DefaultSuccessColorMessage,
			Error:   DefaultErrColorMessage,
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
			log.Debugf("Applying default AuthStrategy %s to command %s", AuthStrategyNone, name)
			command.AuthStrategy = AuthStrategyNone
		}
		if command.Timeout == 0 {
			log.Debugf("Applying default Timeout %d to command %s", DefaultCommandTimeout, name)
			command.Timeout = DefaultCommandTimeout
		} else {
			command.Timeout *= time.Second
			log.Infof("Command timeout for %s is %d", name, command.Timeout)
		}

		// All configured commands are shell type
		command.Type = ShellCommandType

		c.Commands[name] = command // Re-set the command
	}

	return c, nil
}

// Config is the struct used to load MrMeeseeks configuration yaml
type Config struct {
	Messages map[string][]string `yaml:"messages"`
	Commands map[string]Command  `yaml:"commands"`
	Colors   MessageColors       `yaml:"colors"`
}

// Command is the struct that handles a command configuration
type Command struct {
	Cmd          string          `yaml:"command"`
	Args         []string        `yaml:"arguments"`
	Authorized   []string        `yaml:"authorized"`
	AuthStrategy string          `yaml:"auth_strategy"`
	Timeout      time.Duration   `yaml:"timeout"`
	Templates    CommandTemplate `yaml:"templates"`
	Type         int
}

// CommandTemplate is the struct in which the templates used to render a command are kept
type CommandTemplate struct {
	Handshake string `yaml:"on_handshake"`
	Success   string `yaml:"on_success"`
	Failure   string `yaml:"on_failure"`
}

// MessageColors contains the configured reply message colora
type MessageColors struct {
	Info    string `yaml:"info"`
	Success string `yaml:"success"`
	Error   string `yaml:"error"`
}
