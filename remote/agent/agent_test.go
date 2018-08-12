package agent_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/gomeeseeks/meeseeks-box/commands"
	"github.com/gomeeseeks/meeseeks-box/commands/shell"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/mocks"
	"github.com/gomeeseeks/meeseeks-box/remote/agent"
	"github.com/gomeeseeks/meeseeks-box/remote/api"
	"google.golang.org/grpc"
)

type MockServer struct{}

func (m MockServer) RegisterAgent(in *api.AgentConfiguration, agent api.CommandPipeline_RegisterAgentServer) error {
	return nil
}

func (m MockServer) Finish(ctx context.Context, fin *api.CommandFinish) (*api.Empty, error) {
	return &api.Empty{}, nil
}

type MockLogger struct{}

func (MockLogger) Append(writer api.LogWriter_AppendServer) error {
	return nil
}

func (MockLogger) SetError(ctx context.Context, entry *api.ErrorLogEntry) (*api.Empty, error) {
	return &api.Empty{}, nil
}

func TestAgentTalksToServer(t *testing.T) {
	mocks.Must(t, "failed to register commands",
		commands.Register(commands.RegistrationArgs{
			Action: commands.ActionRegister,
			Kind:   commands.KindLocalCommand,
			Commands: []commands.CommandRegistration{
				{
					Name: "remote-echo",
					Cmd: shell.New(meeseeks.CommandOpts{
						Cmd:  "echo",
						Help: meeseeks.NewHelp("echo"),
					}),
				},
			},
		}))
	defer commands.Reset()

	m := MockServer{}
	l := MockLogger{}

	s := grpc.NewServer()
	api.RegisterCommandPipelineServer(s, m)
	api.RegisterLogWriterServer(s, l)

	// Start server
	go func() {
		address, err := net.Listen("tcp", "localhost:9700")
		mocks.Must(t, "wut?", err)
		if err := s.Serve(address); err != nil {
			t.Errorf("failed to start server at %s: %s", address, err)
		}
	}()
	defer s.GracefulStop()

	// Connect agent
	client := agent.New(agent.Configuration{
		GRPCTimeout: 1 * time.Second,
		ServerURL:   "localhost:9700",
		Labels:      map[string]string{"tier": "testing"},
	})
	err := client.Connect()
	mocks.Must(t, "failed to connect to remote server", err)
	client.Shutdown()

}
