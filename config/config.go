package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/gomeeseeks/meeseeks-box/commands"
	"github.com/gomeeseeks/meeseeks-box/commands/shell"

	"github.com/gomeeseeks/meeseeks-box/auth"
	"github.com/gomeeseeks/meeseeks-box/db"

	yaml "gopkg.in/yaml.v2"
)

// Default colors
const (
	DefaultInfoColorMessage    = ""
	DefaultSuccessColorMessage = "good"
	DefaultWarningColorMessage = "warning"
	DefaultErrColorMessage     = "danger"
)

// LoadFile reads the given filename, builds a configuration object and initializes
// all the required subsystems
func LoadFile(filename string) (Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return Config{}, fmt.Errorf("could not open configuration file %s: %s", filename, err)
	}

	cnf, err := New(f)
	if err != nil {
		return cnf, fmt.Errorf("configuration is invalid: %s", err)
	}
	return cnf, nil
}

// LoadConfig loads the configuration in all the dependent subsystems
func LoadConfig(cnf Config) error {
	if err := db.Configure(cnf.Database); err != nil {
		return err
	}
	auth.Configure(cnf.Groups)

	for name, cmd := range cnf.Commands {
		commands.Add(name, shell.New(shell.CommandOpts{
			AllowedGroups: cmd.AllowedGroups,
			Args:          cmd.Args,
			AuthStrategy:  cmd.AuthStrategy,
			Cmd:           cmd.Cmd,
			Help: shell.NewHelp(
				cmd.Help.Summary,
				cmd.Help.Args...),
			Templates: cmd.Templates,
			Timeout:   cmd.Timeout * time.Second,
		}))
	}
	return nil
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

// Command is the struct that handles a command configuration
type Command struct {
	Cmd           string            `yaml:"command"`
	Args          []string          `yaml:"args"`
	AllowedGroups []string          `yaml:"allowed_groups"`
	AuthStrategy  string            `yaml:"auth_strategy"`
	Timeout       time.Duration     `yaml:"timeout"`
	Templates     map[string]string `yaml:"templates"`
	Help          CommandHelp       `yaml:"help"`
	Type          int
}

// CommandHelp is the struct that handles the help of a command
type CommandHelp struct {
	Summary string   `yaml:"summary"`
	Args    []string `yaml:"args"`
}

// MessageColors contains the configured reply message colora
type MessageColors struct {
	Info    string `yaml:"info"`
	Success string `yaml:"success"`
	Error   string `yaml:"error"`
}
