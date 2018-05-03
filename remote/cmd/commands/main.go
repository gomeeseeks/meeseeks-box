package main

import (
// "context"
// "flag"
// "github.com/gomeeseeks/meeseeks-box/remote/api"
// "github.com/sirupsen/logrus"
// "google.golang.org/grpc"
// "sync"
)

func main() {
	// backend := flag.String("address", "localhost:9996", "")
	// flag.Parse()

	// client, err := grpc.Dial(*backend, grpc.WithInsecure())
	// if err != nil {
	// 	logrus.Fatalf("could not connect to client: %s", err)
	// }

	// commands := api.NewCommandPipelineClient(client)
	// logger := api.NewCommandLoggerClient(client)

	// wg := sync.WaitGroup{}
	// wg.Add(1)

	// go func() {
	// 	stream, err := commands.Select(context.Background(), &api.AgentConfiguration{
	// 		Commands: map[string]*api.RemoteCommand{
	// 			"echo": {
	// 				Timeout:      2,
	// 				AuthStrategy: "any",
	// 				Help:         &api.Help{Summary: "prints the text that was sent in"},
	// 			},
	// 		},
	// 		Labels: map[string]string{
	// 			"datacenter": "dc1",
	// 		},
	// 		Token: "my-client-token",
	// 	})
	// 	if err != nil {
	// 		logrus.Fatalf("failed to start listening for commands: %s", err)
	// 	}

	// 	logrus.Infof("waiting on context...")

	// 	for i := 0; i < 2; i++ {
	// 		cmd, err := stream.Recv()
	// 		if err != nil {
	// 			logrus.Fatalf("failed to receive a command: %s", err)
	// 		}
	// 		logrus.Infof("Received command %#v", cmd)

	// 		appender, err := logger.Append(context.Background())
	// 		if err != nil {
	// 			logrus.Fatalf("Could not open an appender stream: %s", err)
	// 		}

	// 		for _, line := range []string{"line1", "line2", "line3", "line4"} {
	// 			l := api.LogEntry{
	// 				JobID: cmd.JobID,
	// 				Line:  line}

	// 			logrus.Info("sending line")
	// 			if err := appender.Send(&l); err != nil {
	// 				logrus.Fatalf("Failed to send log line to backend: %s", err)
	// 			}
	// 			logrus.Infof("line %#v sent", l)

	// 		}
	// 		if _, err := appender.CloseAndRecv(); err != nil {
	// 			logrus.Fatalf("failed to close the channel: %s", err)
	// 		}
	// 		_, err = commands.Finish(context.Background(), &api.CommandFinish{
	// 			JobID:  cmd.JobID,
	// 			Status: "Done",
	// 			Error:  "",
	// 		})
	// 		if err != nil {
	// 			logrus.Fatalf("Failed to finish the command execution %s", err)
	// 		}

	// 	}
	// 	wg.Done()

	// }()

	// wg.Wait()
	// logrus.Infof("Done with it all.")
}
