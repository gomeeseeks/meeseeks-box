package client

import (
	"context"

	"github.com/gomeeseeks/meeseeks-box/remote/api"
	"github.com/sirupsen/logrus"
)

func bla() {
	commands := api.NewCommandPipelineClient(client)
	logger := api.NewCommandLoggerClient(client)

	// Register the commands and start listening on the stream of remote commands
	// On startup it has to send all the commands that the meeseeks knows how to handle (except builtins)
	stream, err := commands.Select(context.Background(), &api.AgentConfiguration{
		Commands: map[string]*api.RemoteCommand{
			"echo": {
				Timeout:      2,
				AuthStrategy: "any",
				Help:         &api.Help{Summary: "prints the text that was sent in"},
			},
		},
		Labels: map[string]string{
			"datacenter": "dc1",
		},
		Token: "my-client-token", // Token should come from the configuration or the environment
	})
	if err != nil {
		logrus.Fatalf("failed to register commands on remote host: %s", err)
	}

	// Get a logAppender 
	appender, err := logger.Append(context.Background())
	if err != nil {
		// Add metric for failing to get a log appender
		logrus.Fatalf("Could not open an appender stream: %s", err)
	}

	ch := make(chan CommandRequest)
	// Get a command (or error)
	go func() {
		for {
			cmd, err := stream.Recv()
			if err != nil {
				logrus.Fatalf("failed to receive a command: %s", err)
			}
			// add a metric to account for remotely received commands
			ch <- cmd
		}
	}

	for cmd := range ch {
		// Execute the command and send all the logs
		cmd.Execute()

		for _, line := range []string{"line1", "line2", "line3", "line4"} {
			l := api.LogEntry{
				JobID: cmd.JobID,
				Line:  line}

			if err := appender.Send(&l); err != nil {
				logrus.Fatalf("Failed to send log line to backend: %s", err)
			}
			// add metric for sent log lines
		}

		if _, err := appender.CloseAndRecv(); err != nil {
			logrus.Fatalf("failed to close the channel: %s", err)
		}

		// Add metric for remotely finished commands
		_, err = commands.Finish(context.Background(), &api.CommandFinish{
			JobID:  cmd.JobID,
			Status: "Done",
			Error:  "",
		})
		if err != nil {
			logrus.Fatalf("Failed to finish the command execution %s", err)
		}
	}

}
