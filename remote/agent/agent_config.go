package agent

import (
	"time"

	"github.com/gomeeseeks/meeseeks-box/commands"
	"github.com/gomeeseeks/meeseeks-box/commands/shell"
	"github.com/gomeeseeks/meeseeks-box/config"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/remote/api"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// Configuration holds the client configuration used to connect to the server
type Configuration struct {
	ServerURL   string
	Options     []grpc.DialOption
	GRPCTimeout time.Duration
	Token       string
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

// GetOptions returns the grpc connection options
func (c *Configuration) GetOptions() []grpc.DialOption {
	if c.Options == nil {
		return []grpc.DialOption{
			grpc.WithInsecure(),
			grpc.WithKeepaliveParams(
				keepalive.ClientParameters{
					Time:                5 * time.Second,
					PermitWithoutStream: true,
					Timeout:             c.GetGRPCTimeout(),
				},
			),
			grpc.WithBackoffMaxDelay(5 * time.Second),
			grpc.WithTimeout(c.GetGRPCTimeout()),
		}
	}
	return c.Options
}

func (c *Configuration) createAgentConfiguration(agentID string) *api.AgentConfiguration {
	return &api.AgentConfiguration{
		Commands: c.createRemoteCommands(),
		Labels:   c.Labels,
		Token:    c.Token,
		AgentID:  agentID,
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
			Help: &api.Help{
				Summary: cmd.Help.Summary,
				Args:    cmd.Help.Args,
			},
		}
	}
	return remoteCommands
}

func (c *Configuration) registerLocalCommands() error {
	cmds := make([]commands.CommandRegistration, 0)
	for name, cmd := range c.Commands {
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
			})})
	}
	logrus.Debugf("registering commands: %#v", cmds)
	return commands.Add(cmds...)
}
