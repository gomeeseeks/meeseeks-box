package agent_test

import (
	"testing"
	"time"

	"github.com/gomeeseeks/meeseeks-box/commands"
	"github.com/gomeeseeks/meeseeks-box/commands/shell"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/mocks"
	"github.com/gomeeseeks/meeseeks-box/remote/agent"
	"github.com/gomeeseeks/meeseeks-box/remote/server"
)

var echoCmd = shell.New(meeseeks.CommandOpts{
	Cmd:  "echo",
	Help: meeseeks.NewHelp("echo"),
})

func TestAgentCanConnectAndRegisterACommand(t *testing.T) {
	commands.Register(commands.RegistrationArgs{
		Action: commands.ActionRegister,
		Kind:   commands.KindLocalCommand,
		Commands: []commands.CommandRegistration{
			commands.CommandRegistration{
				Name: "remote-echo",
				Cmd:  echoCmd,
			},
		},
	})

	s, err := server.New(server.Config{})
	mocks.Must(t, "failed to create grpc server", err)
	defer s.Shutdown()

	go func() {
		mocks.Must(t, "Failed to start server", s.Listen(":9698"))
	}()

	client := agent.New(agent.Configuration{
		GRPCTimeout: 1 * time.Second,
		ServerURL:   "localhost:9698",
		Labels:      map[string]string{"tier": "testing"},
	})
	mocks.Must(t, "failed to connect agent", client.Connect())

	w := make(chan interface{})
	go func() {
		w <- true
		client.Run()
	}()
	<-w
	defer client.Shutdown()
	defer commands.Reset()

	_, ok := commands.Find(&meeseeks.Request{
		Command:     "remote-echo",
		Args:        []string{"hola"},
		IsIM:        false,
		Channel:     "test",
		ChannelID:   "test-id",
		ChannelLink: "test-link",
		UserID:      "user-id",
		Username:    "username",
		UserLink:    "user-link",
	})
	mocks.AssertEquals(t, true, ok)
}

func TestAgentTLSCanConnectAndRegisterACommand(t *testing.T) {
	commands.Register(commands.RegistrationArgs{
		Action: commands.ActionRegister,
		Kind:   commands.KindLocalCommand,
		Commands: []commands.CommandRegistration{
			commands.CommandRegistration{
				Name: "remote-echo",
				Cmd:  echoCmd,
			},
		},
	})

	s, err := server.New(server.Config{
		SecurityMode: "tls",
		CertPath:     "../../config/test-fixtures/cert.pem",
		KeyPath:      "../../config/test-fixtures/key.pem",
	})
	mocks.Must(t, "failed to create grpc server", err)
	defer s.Shutdown()
	defer commands.Reset()

	go func() {
		mocks.Must(t, "Failed to start server", s.Listen(":9698"))
	}()

	client := agent.New(agent.Configuration{
		GRPCTimeout:  1 * time.Second,
		ServerURL:    "localhost:9698",
		SecurityMode: "tls",
		CertPath:     "../../config/test-fixtures/cert.pem",
		Labels:       map[string]string{"tier": "testing"},
	})
	mocks.Must(t, "failed to connect agent", client.Connect())

	w := make(chan interface{})
	go func() {
		w <- true
		client.Run()
	}()
	<-w
	defer client.Shutdown()

	_, ok := commands.Find(&meeseeks.Request{
		Command:     "remote-echo",
		Args:        []string{"hola"},
		IsIM:        false,
		Channel:     "test",
		ChannelID:   "test-id",
		ChannelLink: "test-link",
		UserID:      "user-id",
		Username:    "username",
		UserLink:    "user-link",
	})
	mocks.AssertEquals(t, true, ok)
}
