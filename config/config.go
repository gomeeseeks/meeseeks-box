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

// New parses the configuration from a reader into an object and returns it
func New(r io.Reader) (Config, error) {
	c := Config{}

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
	Cmd          string   `yaml:"command"`
	Args         []string `yaml:"arguments"`
	Authorized   []string `yaml:"authorized"`
	AuthStrategy string   `yaml:"auth_strategy"`
	Timeout      int      `yaml:"timeout"`
}
