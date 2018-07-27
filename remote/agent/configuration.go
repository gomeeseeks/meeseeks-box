package agent

import (
	"time"

	"github.com/gomeeseeks/meeseeks-box/commands"
	"github.com/gomeeseeks/meeseeks-box/commands/shell"
	"github.com/gomeeseeks/meeseeks-box/config"
	"github.com/gomeeseeks/meeseeks-box/remote/api"

	"google.golang.org/grpc"
)

// Configuration holds the client configuration used to connect to the server
type Configuration struct {
	ServerURL   string
	Options     grpc.DialOption // grpc.WithInsecure()
	GRPCTimeout time.Duration
	Token       string
	Pool        int
	Labels      map[string]string
	Commands    map[string]config.Command
}

// GetGRPCTimeout returns the configured timeout or a default of 10 seconds
func (c *Configuration) GetGRPCTimeout() time.Duration {
	if c.GRPCTimeout == 0 {
		return 10 * time.Second
	}
	return c.GRPCTimeout
}

func (c *Configuration) createAgentConfiguration() *api.AgentConfiguration {
	return &api.AgentConfiguration{
		Commands: c.createRemoteCommands(),
		Labels:   c.Labels,
		Token:    c.Token,
	}
}

func (c *Configuration) createRemoteCommands() map[string]*api.RemoteCommand {
	remoteCommands := make(map[string]*api.RemoteCommand, len(c.Commands))
	for name, cmd := range c.Commands {
		remoteCommands[name] = &api.RemoteCommand{
			Timeout:         int64(cmd.Timeout),
			AuthStrategy:    cmd.AuthStrategy,
			AllowedGroups:   cmd.AllowedGroups,
			ChannelStrategy: cmd.ChannelStrategy,
			AllowedChannels: cmd.AllowedChannels,
			Templates:       cmd.Templates,
			Help: &api.Help{
				Summary: cmd.Help.Summary,
				Args:    cmd.Help.Args,
			},
		}
	}
	return remoteCommands
}

func (c *Configuration) registerLocalCommands() {
	for name, cmd := range c.Commands {
		commands.Add(name, shell.New(shell.CommandOpts{
			AuthStrategy:    cmd.AuthStrategy,
			AllowedGroups:   cmd.AllowedGroups,
			ChannelStrategy: cmd.ChannelStrategy,
			AllowedChannels: cmd.AllowedChannels,
			Args:            cmd.Args,
			HasHandshake:    !cmd.NoHandshake,
			Cmd:             cmd.Cmd,
			Help: shell.NewHelp(
				cmd.Help.Summary,
				cmd.Help.Args...),
			Templates: cmd.Templates,
			Timeout:   cmd.Timeout * time.Second,
		}))
	}

}
