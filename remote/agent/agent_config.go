package agent

import (
	"time"

	"gitlab.com/yakshaving.art/meeseeks-box/commands"
	"gitlab.com/yakshaving.art/meeseeks-box/remote/api"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// SecurityModeTLS means TLS security mode with server cert
const SecurityModeTLS = "tls"

// Configuration holds the client configuration used to connect to the server
type Configuration struct {
	ServerURL   string
	GRPCTimeout time.Duration

	SecurityMode string
	CertPath     string

	Token  string
	Labels map[string]string
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
	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(
			keepalive.ClientParameters{
				Time:                5 * time.Second,
				PermitWithoutStream: true,
				Timeout:             c.GetGRPCTimeout(),
			},
		),
		grpc.WithBackoffMaxDelay(5 * time.Second),
		grpc.WithTimeout(c.GetGRPCTimeout()),
		grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
	}
	switch c.SecurityMode {
	case SecurityModeTLS:
		creds, err := credentials.NewClientTLSFromFile(c.CertPath, "")
		if err != nil {
			logrus.Fatalf("could not load server cert: %s", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))

	default:
		logrus.Warnf("using insecure client mode")
		opts = append(opts, grpc.WithInsecure())
	}
	return opts
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
	remoteCommands := make(map[string]*api.RemoteCommand, 0)
	for name, cmd := range commands.All() {
		remoteCommands[name] = &api.RemoteCommand{
			Timeout:         cmd.GetTimeout().Nanoseconds(),
			AuthStrategy:    cmd.GetAuthStrategy(),
			AllowedGroups:   cmd.GetAllowedGroups(),
			ChannelStrategy: cmd.GetChannelStrategy(),
			AllowedChannels: cmd.GetAllowedChannels(),
			Help: &api.Help{
				Summary: cmd.GetHelp().GetSummary(),
				Args:    cmd.GetHelp().GetArgs(),
			},
		}
	}
	return remoteCommands
}
