package config

import (
	"fmt"
	"io"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Authorization Strategies determine who has access to what
const (
	AuthStrategyAny      = "any"
	AuthStrategyUserList = "userlist"
)

var (
	defaultHandshake    = []string{"I'm Mr Meeseeks! look at me!", "Mr Meeseeks!", "Uuuuh, yeah! can do!", "Can doo...", "Uuuuh, ok!"}
	defaultSuccess      = []string{"All done!", "Mr Meeseeks", "Uuuuh, nice!"}
	defaultFailed       = []string{"Uuuh!, no, it failed"}
	defaultUnauthorized = []string{"Uuuh! no! no soup for you"}
)

// New parses the configuration from a reader into an object and returns it
func New(r io.Reader) (Config, error) {
	c := Config{
		Messages: map[string][]string{
			"handshake":    defaultHandshake,
			"success":      defaultSuccess,
			"failed":       defaultFailed,
			"unauthorized": defaultUnauthorized,
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

	for _, command := range c.Commands {
		if command.AuthStrategy == "" {
			command.AuthStrategy = AuthStrategyAny
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
	Timeout      int             `yaml:"timeout"`
	Templates    CommandTemplate `yaml:"templates"`
}

// CommandTemplate is the struct in which the templates used to render a command are kept
type CommandTemplate struct {
	Handshake string `yaml:"on_handshake"`
	Success   string `yaml:"on_success"`
	Failure   string `yaml:"on_failure"`
}
