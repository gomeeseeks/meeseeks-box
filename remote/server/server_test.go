package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/gomeeseeks/meeseeks-box/commands"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/mocks"
	"github.com/gomeeseeks/meeseeks-box/remote/api"
	"github.com/gomeeseeks/meeseeks-box/remote/server"

	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
)

func TestAgentCanConnect(t *testing.T) {
	mocks.WithTmpDB(func(_ string) {

		s, err := server.New(server.Config{})
		mocks.Must(t, "failed to create grpc server", err)
		defer s.Shutdown()

		c := make(chan interface{})
		go func() {
			c <- true
			mocks.Must(t, "Failed to start server", s.Listen(":9699"))
		}()

		<-c
		time.Sleep(1 * time.Millisecond)

		client, err := grpc.Dial("localhost:9699", grpc.WithInsecure())
		mocks.Must(t, "could not create grpc client", err)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		cmdClient := api.NewCommandPipelineClient(client)
		pipeline, err := cmdClient.RegisterAgent(ctx, &api.AgentConfiguration{
			AgentID: "agentID1",
			Token:   "mytoken",
			Labels:  map[string]string{},
			Commands: map[string]*api.RemoteCommand{
				"echo": {
					AllowedChannels: []string{},
					AllowedGroups:   []string{},
					AuthStrategy:    "any",
					ChannelStrategy: "any",
					Timeout:         10,
				},
			},
		})
		mocks.Must(t, "could not register client", err)
		logrus.Infof("done registering agent, apparently")

		time.Sleep(10 * time.Millisecond)

		go func() {
			logrus.Infof("commands: %#v", commands.All())
			cmd, ok := commands.Find(&meeseeks.Request{
				Command: "echo",
			})
			logrus.Infof("fetching command echo: %b", ok)
			mocks.AssertEquals(t, true, ok)

			cmd.Execute(ctx, meeseeks.Job{
				ID: 27,
				Request: meeseeks.Request{
					Command: "echo",
				},
				Status: meeseeks.JobRunningStatus,
			})
		}()

		cmdReq, err := pipeline.Recv()
		mocks.Must(t, "failed receiving command requests", err)

		mocks.AssertEquals(t, uint64(27), cmdReq.JobID)
		mocks.AssertEquals(t, "echo", cmdReq.Command)

		logClient := api.NewLogWriterClient(client)
		appender, err := logClient.Append(ctx)
		mocks.Must(t, "could not create log appender", err)

		mocks.Must(t, "could not send log line", appender.Send(&api.LogEntry{JobID: cmdReq.JobID, Line: "log line 1"}))

		cmdClient.Finish(ctx, &api.CommandFinish{
			AgentID: "agentID1",
			Content: "done",
			JobID:   cmdReq.JobID,
		})
	})
}

func TestUsingTLS(t *testing.T) {
	s, err := server.New(server.Config{
		SecurityMode: server.SecurityModeTLS,
		CertPath:     "../../config/test-fixtures/cert.pem",
		KeyPath:      "../../config/test-fixtures/key.pem",
	})
	mocks.Must(t, "failed to create grpc server", err)

	defer s.Shutdown()

	c := make(chan interface{})
	go func() {
		c <- true
		mocks.Must(t, "Failed to start server", s.Listen(":9700"))
	}()

	<-c
	time.Sleep(1 * time.Millisecond)
}
