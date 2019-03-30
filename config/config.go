package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/gomeeseeks/meeseeks-box/commands"
	"github.com/gomeeseeks/meeseeks-box/commands/shell"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"

	"github.com/gomeeseeks/meeseeks-box/auth"
	"github.com/gomeeseeks/meeseeks-box/persistence/db"
	"github.com/gomeeseeks/meeseeks-box/text/formatter"

	yaml "gopkg.in/yaml.v2"
)

// ReadFile reads the given filename and returns a configuration object
func ReadFile(filename string) (Config, error) {
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

// LoadConfiguration loads the configuration in all the dependent subsystems
func LoadConfiguration(cnf Config) error {
	if err := db.Configure(cnf.Database); err != nil {
		return fmt.Errorf("could not configure database: %s", err)
	}

	cmds := make([]commands.CommandRegistration, 0)
	for name, cmd := range cnf.Commands {
		cmds = append(cmds, commands.CommandRegistration{
			Name: name,
			Cmd: shell.New(meeseeks.CommandOpts{
				AuthStrategy:    cmd.AuthStrategy,
				AllowedGroups:   cmd.AllowedGroups,
				ChannelStrategy: cmd.ChannelStrategy,
				AllowedChannels: cmd.AllowedChannels,
				Args:            cmd.Args,
				Handshake:       !cmd.NoHandshake,
				Cmd:             cmd.Cmd,
				Help: meeseeks.NewHelp(
					cmd.Help.Summary,
					cmd.Help.Args...),
				Timeout: cmd.Timeout * time.Second,
			}),
		})
	}
	if err := commands.Register(commands.RegistrationArgs{
		Kind:     commands.KindLocalCommand,
		Action:   commands.ActionRegister,
		Commands: cmds,
	}); err != nil {
		return fmt.Errorf("could not load commands: %s", err)
	}

	auth.Configure(cnf.Groups)
	formatter.Configure(cnf.Format)

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
		Format: formatter.FormatConfig{
			Colors: formatter.MessageColors{
				Info:    formatter.DefaultInfoColorMessage,
				Success: formatter.DefaultSuccessColorMessage,
				Error:   formatter.DefaultErrColorMessage,
			},
			ReplyStyle: map[string]string{},
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
	Database db.DatabaseConfig      `yaml:"database"`
	Commands map[string]Command     `yaml:"commands"`
	Groups   map[string][]string    `yaml:"groups"`
	Pool     int                    `yaml:"pool"`
	Format   formatter.FormatConfig `yaml:"format"`
}

// Command is the struct that handles a command configuration
type Command struct {
	Cmd             string        `yaml:"command"`
	Args            []string      `yaml:"args"`
	AllowedGroups   []string      `yaml:"allowed_groups"`
	AuthStrategy    string        `yaml:"auth_strategy"`
	ChannelStrategy string        `yaml:"channel_strategy"`
	AllowedChannels []string      `yaml:"allowed_channels"`
	NoHandshake     bool          `yaml:"no_handshake"`
	Timeout         time.Duration `yaml:"timeout"`
	Help            CommandHelp   `yaml:"help"`
}

// CommandHelp is the struct that handles the help of a command
type CommandHelp struct {
	Summary string   `yaml:"summary"`
	Args    []string `yaml:"args"`
}
