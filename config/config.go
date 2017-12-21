package config

import (
	"fmt"
	"io"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

const (
	ActionSelect = "select"
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

	return c, nil
}

// Config is the struct used to load MrMeeseeks configuration yaml
type Config struct {
	Messages map[string][]string `yaml:"messages"`
	Commands map[string]Command  `yaml:"commands"`
}

// Command is the struct that handles a command configuration
type Command struct {
	Cmd        string         `yaml:"command"`
	Args       []string       `yaml:"arguments"`
	Authorized []string       `yaml:"authorized"`
	Action     string         `yaml:"action"`
	Options    CommandOptions `yaml:"options,omitempty"`
}

// CommandOptions
type CommandOptions struct {
	Message string            `yaml:"message"`
	Values  map[string]string `yaml:"values"`
}
