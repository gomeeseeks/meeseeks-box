package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/pcarranza/meeseeks-box/auth"
	"github.com/pcarranza/meeseeks-box/command"
	"github.com/pcarranza/meeseeks-box/db"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// AdminGroup is the default admin group used by builtin commands
const AdminGroup = "admin"

// Default colors
const (
	DefaultInfoColorMessage    = ""
	DefaultSuccessColorMessage = "good"
	DefaultWarningColorMessage = "warning"
	DefaultErrColorMessage     = "danger"
)

// Command types
const (
	BuiltinCommandType = iota
	ShellCommandType
	RemoteCommandType
)

// Load reads the given filename, builds a configuration object and initializes
// all the required subsystems
func Load(filename string) (Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return Config{}, fmt.Errorf("could not open configuration file %s: %s", filename, err)
	}

	cnf, err := New(f)
	if err != nil {
		return cnf, fmt.Errorf("configuration is invalid: %s", err)
	}

	db.Configure(cnf.Database)
	auth.Configure(cnf.Groups)

	return cnf, nil
}

// New parses the configuration from a reader into an object and returns it
func New(r io.Reader) (Config, error) {
	c := Config{
		Database: db.DatabaseConfig{
			Path:    "meeseeks.db",
			Mode:    0600,
			Timeout: 2 * time.Second,
		},
		Colors: MessageColors{
			Info:    DefaultInfoColorMessage,
			Success: DefaultSuccessColorMessage,
			Error:   DefaultErrColorMessage,
		},
		Pool: 20,
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return c, fmt.Errorf("could not read configuration: %s", err)
	}

	err = yaml.Unmarshal(b, &c)
	if err != nil {
		return c, fmt.Errorf("could not parse configuration: %s", err)
	}

	for name, cmd := range c.Commands {
		if cmd.AuthStrategy == "" {
			log.Debugf("Applying default AuthStrategy %s to command %s", auth.AuthStrategyNone, name)
			cmd.AuthStrategy = auth.AuthStrategyNone
		}
		if cmd.Timeout == 0 {
			log.Debugf("Applying default Timeout %d sec to command %s", command.DefaultCommandTimeout/time.Second, name)
			cmd.Timeout = command.DefaultCommandTimeout
		} else {
			cmd.Timeout *= time.Second
			log.Infof("Command timeout for %s is %d seconds", name, cmd.Timeout/time.Second)
		}

		// All configured commands are shell type
		cmd.Type = ShellCommandType

		c.Commands[name] = cmd // Re-set the command
	}

	return c, nil
}

// Config is the struct used to load MrMeeseeks configuration yaml
type Config struct {
	Database db.DatabaseConfig   `yaml:"database"`
	Messages map[string][]string `yaml:"messages"`
	Commands map[string]Command  `yaml:"commands"`
	Colors   MessageColors       `yaml:"colors"`
	Groups   map[string][]string `yaml:"groups"`
	Pool     int                 `yaml:"pool"`
}

// CommandConfig is the struct that handles a command configuration
type Command struct {
	Cmd           string            `yaml:"command"`
	Args          []string          `yaml:"args"`
	AllowedGroups []string          `yaml:"allowed_groups"`
	AuthStrategy  string            `yaml:"auth_strategy"`
	Timeout       time.Duration     `yaml:"timeout"`
	Templates     map[string]string `yaml:"templates"`
	Help          string            `yaml:"help"`
	Type          int
}

// MessageColors contains the configured reply message colora
type MessageColors struct {
	Info    string `yaml:"info"`
	Success string `yaml:"success"`
	Error   string `yaml:"error"`
}
