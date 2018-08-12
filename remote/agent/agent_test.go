package agent_test

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/gomeeseeks/meeseeks-box/commands"
	"github.com/gomeeseeks/meeseeks-box/commands/shell"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/mocks"
	"github.com/gomeeseeks/meeseeks-box/remote/agent"
	"github.com/gomeeseeks/meeseeks-box/remote/api"

	"github.com/onrik/logrus/filename"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var wg = sync.WaitGroup{}
var ch = make(chan api.CommandFinish)

func init() {
	logrus.AddHook(filename.NewHook())
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.DebugLevel)
}

var echoCmd = shell.New(meeseeks.CommandOpts{
	Cmd:  "echo",
	Help: meeseeks.NewHelp("echo"),
})

type MockServer struct{}

func (m MockServer) RegisterAgent(in *api.AgentConfiguration, agent api.CommandPipeline_RegisterAgentServer) error {
	logrus.Infof("mock server: sending command to agent")
	err := agent.Send(&api.CommandRequest{
		JobID:    1,
		Command:  "echo",
		Args:     []string{"something", "something"},
		Channel:  "channel",
		Username: "someone",
	})
	logrus.Infof("mock server: command sent to agent")
	if err != nil {
		return fmt.Errorf("failed to send command request: %s", err)
	}

	wg.Add(1)
	wg.Wait()

	logrus.Infof("mock server: done, exiting")
	return nil
}

func (m MockServer) Finish(ctx context.Context, fin *api.CommandFinish) (*api.Empty, error) {
	logrus.Infof("mock server: storing finished command")

	ch <- *fin
	wg.Done()
	return &api.Empty{}, nil
}

type MockLogger struct {
	logs []string
}

func (l MockLogger) Append(writer api.LogWriter_AppendServer) error {

	logrus.Infof("mock server: append to logs")
	log, err := writer.Recv()
	if err != nil {
		return fmt.Errorf("error when appending to log: %s", err)
	}
	l.logs = append(l.logs, fmt.Sprintf("%d-%s", log.GetJobID(), log.GetLine()))
	logrus.Infof("mock server: done appending to logs")

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
					Name: "echo",
					Cmd:  echoCmd,
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
		logrus.Infof("agent test: starting server")
		address, err := net.Listen("tcp", "localhost:9700")
		mocks.Must(t, "wut?", err)
		if err := s.Serve(address); err != nil {
			t.Errorf("failed to start server at %s: %s", address, err)
		}
		logrus.Infof("agent test: stopped server")
	}()
	defer s.GracefulStop()

	// Connect agent
	client := agent.New(agent.Configuration{
		GRPCTimeout: 10 * time.Second,
		ServerURL:   "localhost:9700",
		Labels:      map[string]string{"tier": "testing"},
	})
	logrus.Infof("agent test: connecting client")
	err := client.Connect()
	mocks.Must(t, "failed to connect to remote server", err)
	logrus.Infof("agent test: client connected")

	go func() {
		logrus.Infof("agent test: running client")
		client.Run()
		logrus.Infof("agent test: client stopped")
	}()

	finished := <-ch

	logrus.Infof("agent test: shutting down client")
	client.Shutdown()

	mocks.AssertEquals(t, "something something\n", finished.GetContent())
	mocks.AssertEquals(t, uint64(1), finished.GetJobID())
}
